/*
Copyright © 2020 Hoshea Jiang <hoshea@apache.org>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"os"
	"testing"
)

func TestGetFileExtension(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "txt extension",
			args: args{
				filename: ".abc.txt",
			},
			want: "txt",
		},
		{
			name: "no file extensions",
			args: args{
				filename: ".abc.txt.",
			},
			want: "",
		},
		{
			name: "no file extensions",
			args: args{
				filename: "lkjsdl",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtension(tt.args.filename); got != tt.want {
				t.Errorf("GetFileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cleanPathPrefixes(t *testing.T) {
	type args struct {
		path     string
		prefixes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				path:     "test/exclude_test",
				prefixes: []string{"test", string(os.PathSeparator)},
			},
			want: "exclude_test",
		},
		{
			name: "",
			args: args{
				path:     "test/exclude_test/directories",
				prefixes: []string{"test", string(os.PathSeparator)},
			},
			want: "exclude_test/directories",
		},
		{
			name: "",
			args: args{
				path:     "./.git/",
				prefixes: []string{".", string(os.PathSeparator)},
			},
			want: ".git/",
		},
		{
			name: "",
			args: args{
				path:     "test/exclude_test/directories",
				prefixes: []string{"test/", string(os.PathSeparator)},
			},
			want: "exclude_test/directories",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CleanPathPrefixes(tt.args.path, tt.args.prefixes); got != tt.want {
				t.Errorf("CleanPathPrefixes() = %v, want %v", got, tt.want)
			}
		})
	}
}
