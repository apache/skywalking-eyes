//
// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package deps

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
)

const (
	pomFileName = "pom.xml"
)

func (mvn *mavenTool) NewProject() (*Project, error) {
	project := &Project{modules: map[string]*pomFileWrapper{}, mavenTool: mvn}

	// load the root module of the project
	pom, err := newPomfile(".", pomFileName)
	if err != nil {
		return nil, err
	}

	err = project.LoadSubModules(pom)
	if err != nil {
		return nil, err
	}

	project.ConstructDependentGraph()

	return project, nil
}

type Project struct {
	*mavenTool
	modules map[string]*pomFileWrapper
}

func (project *Project) LoadSubModules(pom *pomFileWrapper) error {
	stack := []*pomFileWrapper{pom}

	for len(stack) > 0 {
		p := stack[0]
		stack = stack[1:]
		project.addModule(p)

		modules := p.GetAllModules()
		for _, module := range modules {
			path := filepath.Join(p.Path, module)

			subPom, err := newPomfile(path, pomFileName)
			if err != nil {
				return err
			}

			if _, have := project.modules[subPom.Name]; !have {
				stack = append(stack, subPom)
			}
		}
	}

	return nil
}

func (project *Project) addModule(pom *pomFileWrapper) {
	if _, have := project.modules[pom.Name]; !have {
		project.modules[pom.Name] = pom
	}
}

func (project *Project) ConstructDependentGraph() {
	for k := range project.modules {
		pom := project.modules[k]

		if pom.Parent != nil {
			parentName := pom.ParentName()
			if parent, have := project.modules[parentName]; have {
				pom.haveParentInProject = have
				pom.Uncles[parentName] = parent
			}
		}

		uncles := project.GetAllUncles(pom)
		for _, uncleName := range uncles {
			if uncle, have := project.modules[uncleName]; have {
				pom.Uncles[uncleName] = uncle
			}
		}
	}
}

func (project *Project) GetAllUncles(pom *pomFileWrapper) []string {
	uncles := []string{}

	for k := range pom.cacheDeps {
		if _, have := project.modules[k]; have {
			uncles = append(uncles, k)
		}
	}

	for k := range pom.cacheDepsManager {
		if _, have := project.modules[k]; have {
			uncles = append(uncles, k)
		}
	}

	return uncles
}

var (
	cpuNumber = runtime.NumCPU()
)

type rst struct {
	name string
	deps []*DependencyWrapper
}

func (project *Project) LoadDependencies() ([]*DependencyWrapper, error) {
	modules := project.splitModules()

	wg := &waitGroup{
		v:         new(int64),
		WaitGroup: &sync.WaitGroup{},
	}

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().Unix()))
	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	errCh := make(chan error)
	rstCh := make(chan *rst)
	moduleCh := make(chan *pomFileWrapper)

	// start worker
	for i := 0; i < cpuNumber; i++ {
		dir := filepath.Join(tmpDir, fmt.Sprintf("%d_%d", i, time.Now().Unix()))
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			close(moduleCh)
			return nil, err
		}
		wg.Add()
		go project.worker(wg, dir, moduleCh, rstCh, errCh)
	}

	// start collector
	depsMap := make(map[string]*DependencyWrapper)
	go project.collector(wg, rstCh, depsMap, len(modules))

	// start sender
	wg.Add()
	go project.sender(wg, modules, moduleCh)

	wg.Wait()

	// waiting for collector
	wg.Add()
	close(rstCh)
	wg.Wait()
	if wg.err != nil {
		return nil, wg.err
	}

	keys := make([]string, 0, len(depsMap))
	for k := range depsMap {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	allDeps := make([]*DependencyWrapper, 0, len(depsMap))
	for _, key := range keys {
		allDeps = append(allDeps, depsMap[key])
	}

	return allDeps, nil
}

type waitGroup struct {
	*sync.WaitGroup
	v       *int64
	cnt     int64
	err     error
	errOnce sync.Once
}

func (wg *waitGroup) Done() {
	wg.cnt--
	wg.WaitGroup.Done()
	atomic.AddInt64(wg.v, -1)
}

func (wg *waitGroup) Add() {
	wg.cnt++
	wg.WaitGroup.Add(1)
	atomic.AddInt64(wg.v, 1)
}

func (wg *waitGroup) Running() bool {
	v := atomic.LoadInt64(wg.v)
	return v >= 0
}

func (wg *waitGroup) Err(err error) {
	wg.errOnce.Do(func() {
		wg.err = err
		atomic.AddInt64(wg.v, -2*wg.cnt)
	})
}

func (project *Project) sender(wg *waitGroup, modules []*pomFileWrapper, moduleCh chan<- *pomFileWrapper) {
	defer wg.Done()
	for _, module := range modules {
		moduleCh <- module
		if !wg.Running() {
			break
		}
	}
	close(moduleCh)
}

func (project *Project) worker(wg *waitGroup, tmpDir string, moduleCh <-chan *pomFileWrapper, rstCh chan<- *rst, errCh chan<- error) {
	defer wg.Done()
	for module := range moduleCh {
		if !wg.Running() {
			return
		}

		tmpFile, err := os.OpenFile(filepath.Join(tmpDir, pomFileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			wg.Err(err)
			continue
		}

		_, err = tmpFile.Write(module.Encode())
		if err != nil {
			wg.Err(err)
			continue
		}

		deps, err := project.loadDependencies(tmpFile.Name())
		if err != nil {
			wg.Err(err)
			continue
		}
		rstCh <- &rst{module.Name, deps}

		err = tmpFile.Close()
		if err != nil {
			wg.Err(err)
			continue
		}
	}
}

func (project *Project) collector(wg *waitGroup, rstCh <-chan *rst, depsMap map[string]*DependencyWrapper, all int) {
	defer wg.Done()
	cnt := 0
	for rst := range rstCh {
		cnt++
		for _, dep := range rst.deps {
			if _, have := depsMap[dep.DepName()]; have {
				continue
			}

			if _, have := project.modules[dep.DepName()]; have {
				continue
			}

			depsMap[dep.DepName()] = dep
		}
		logger.Log.Debugf("module %s successfully processed: %d/%d", rst.name, cnt, all)
	}
}

func (project *Project) splitModules() []*pomFileWrapper {
	project.splitParent()

	globalVisitedPom := map[string]bool{}
	modules := make([]*pomFileWrapper, 0, len(project.modules))
	for {
		var pom *pomFileWrapper
		cnt := 0
		for k := range project.modules {
			if globalVisitedPom[k] {
				cnt++
				continue
			}

			pom = project.modules[k]
			break
		}

		if pom == nil {
			break
		}

		visitedUncles := map[string]bool{}

		stack := []string{}
		for uncleName := range pom.Uncles {
			stack = append(stack, uncleName)
			visitedUncles[uncleName] = true
		}

		for len(stack) > 0 {
			key := stack[0]
			stack = stack[1:]

			uncle := project.modules[key]
			pom.Inherit(uncle)

			for uncleName := range uncle.Uncles {
				if visitedUncles[uncleName] {
					continue
				}
				stack = append(stack, uncleName)
				visitedUncles[uncleName] = true
			}
		}

		pom.Uncles = nil
		modules = append(modules, pom)
		globalVisitedPom[pom.Name] = true
	}

	project.Flush()
	return modules
}

func (project *Project) splitParent() {
	visitedPom := map[string]bool{}

	for k := range project.modules {
		pom := project.modules[k]
		if !pom.haveParentInProject {
			visitedPom[pom.Name] = true
			pom.externalParent = pom.Parent
		}
	}

	for len(visitedPom) < len(project.modules) {
		for k := range project.modules {
			if visitedPom[k] {
				continue
			}

			pom := project.modules[k]
			parentName := pom.ParentName()

			if !visitedPom[parentName] {
				continue
			}

			parent := project.modules[parentName]
			pom.Inherit(parent)
			pom.externalParent = parent.externalParent
			visitedPom[pom.Name] = true
		}
	}
}

func (project *Project) Flush() {
	for _, pom := range project.modules {
		project.flush(pom)
	}
}

func (project *Project) flush(pom *pomFileWrapper) {
	pom.Clear()

	pom.Parent = pom.externalParent

	deps := new(Dependencies)
loop:
	for key, dep := range pom.cacheDeps {
		if _, have := project.modules[key]; have {
			continue
		}

		for _, skip := range skipScppe {
			if dep.Scope == skip {
				continue loop
			}
		}

		deps.Value = append(deps.Value, dep)
	}
	pom.PomFile.Dependencies = deps

	depsManager := new(DependencyManagement)
	for key, dep := range pom.cacheDepsManager {
		if _, have := project.modules[key]; have {
			continue
		}
		depsManager.Value.Value = append(depsManager.Value.Value, dep)
	}
	pom.PomFile.DependencyManagement = depsManager

	pom.PomFile.Properties = &Properties{m: pom.cacheProperties}
}

func DepNameNoVersion(dep *Dependency) string {
	return fmt.Sprintf("%v/%v", dep.GroupID, dep.ArtifactID)
}

func newPomfile(path, fileName string) (*pomFileWrapper, error) {
	var file *os.File
	var err error
	file, err = os.OpenFile(filepath.Join(path, fileName), os.O_RDONLY, 0400)

	if err != nil {
		return nil, err
	}

	rawPom, err := DecodePomFile(file)
	if err != nil {
		return nil, err
	}

	pom := &pomFileWrapper{
		PomFile:          rawPom,
		Path:             path,
		cacheDeps:        map[string]*Dependency{},
		cacheDepsManager: map[string]*Dependency{},
		cacheProperties:  map[string]string{},
		Uncles:           map[string]*pomFileWrapper{},
	}
	pom.Name = pom.GetName()

	if pom.Dependencies != nil {
		for _, dep := range pom.Dependencies.Value {
			pom.cacheDeps[DepNameNoVersion(dep)] = dep
		}
	}

	if pom.DependencyManagement != nil {
		for _, dep := range pom.DependencyManagement.Value.Value {
			pom.cacheDepsManager[DepNameNoVersion(dep)] = dep
		}
	}

	if pom.PomFile.Properties != nil {
		for k, v := range pom.PomFile.Properties.m {
			pom.cacheProperties[k] = v
		}
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return pom, nil
}

// pomFileWrapper is a wrapper of the PomFile
type pomFileWrapper struct {
	*PomFile
	Path                string
	Name                string
	haveParentInProject bool
	externalParent      *Parent
	cacheDeps           map[string]*Dependency
	cacheDepsManager    map[string]*Dependency
	cacheProperties     map[string]string
	Uncles              map[string]*pomFileWrapper
}

func (pom *pomFileWrapper) GetAllModules() []string {
	modules := []string{}

	if pom.PomFile.Modules != nil {
		for _, module := range pom.PomFile.Modules.Values {
			modules = append(modules, module.Value)
		}
	}

	if pom.PomFile.Profiles != nil {
		for _, profile := range pom.Profiles.Values {
			for _, module := range profile.Modules.Values {
				modules = append(modules, module.Value)
			}
		}
	}

	return modules
}

func (pom *pomFileWrapper) GetName() string {
	gid := pom.GroupID
	if gid == "" && pom.PomFile.Parent != nil {
		gid = pom.PomFile.Parent.GroupID
	}

	aid := pom.ArtifactID
	if aid == "" && pom.PomFile.Parent != nil {
		aid = pom.PomFile.Parent.ArtifactID
	}

	version := pom.Version
	if version == "" && pom.PomFile.Parent != nil {
		version = pom.PomFile.Parent.Version
	}

	pom.GroupID, pom.ArtifactID, pom.Version = gid, aid, version

	return fmt.Sprintf("%v/%v", gid, aid)
}

func (pom *pomFileWrapper) Clear() {
	pom.PomFile.Modules = nil
	pom.PomFile.Profiles = nil
}

func (pom *pomFileWrapper) Inherit(parent *pomFileWrapper) {
	pom.InheritDeps(parent)
	pom.InheritDepsManager(parent)
	pom.InheritProperties(parent)
}

func (pom *pomFileWrapper) InheritDeps(parent *pomFileWrapper) {
	for k, v := range parent.cacheDeps {
		if _, have := pom.cacheDeps[k]; have {
			continue
		}
		pom.cacheDeps[k] = v
	}
}

func (pom *pomFileWrapper) InheritDepsManager(parent *pomFileWrapper) {
	for k, v := range parent.cacheDepsManager {
		if _, have := pom.cacheDepsManager[k]; have {
			continue
		}
		pom.cacheDepsManager[k] = v
	}
}

func (pom *pomFileWrapper) InheritProperties(parent *pomFileWrapper) {
	for k, v := range parent.cacheProperties {
		if _, have := pom.cacheProperties[k]; have {
			continue
		}
		pom.cacheProperties[k] = v
	}
}

func (pom *pomFileWrapper) ParentName() string {
	return fmt.Sprintf("%v/%v", pom.Parent.GroupID, pom.Parent.ArtifactID)
}

func (pom *pomFileWrapper) HaveLicenses() bool {
	return pom.Licenses != nil && len(pom.Licenses.Values) > 0
}

func (pom *pomFileWrapper) AllLicenses() string {
	licenses := []string{}
	for _, l := range pom.Licenses.Values {
		if l.Name != "" {
			licenses = append(licenses, l.Name)
			continue
		}
		licenses = append(licenses, l.URL)
	}
	return strings.Join(licenses, ", ")
}

func (pom *pomFileWrapper) Raw() string {
	buf := bytes.NewBuffer(nil)
	enc := xml.NewEncoder(buf)
	err := enc.Encode(pom.Licenses)
	if err != nil {
		return ""
	}
	return buf.String()
}
