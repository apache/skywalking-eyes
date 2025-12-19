# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

$LOAD_PATH.unshift(File.expand_path('../lib', __FILE__))

require 'citrus/version'

Gem::Specification.new do |s|
  s.name = 'citrus'
  s.version = Citrus.version
  s.date = Time.now.strftime('%Y-%m-%d')

  s.summary = 'Parsing Expressions for Ruby'
  s.description = 'Parsing Expressions for Ruby'

  s.author = 'Michael Jackson'
  s.email = 'maintainer@example.com'

  s.require_paths = %w< lib >

  s.files = Dir['benchmark/**'] +
    Dir['doc/**'] +
    Dir['extras/**'] +
    Dir['lib/**/*.rb'] +
    Dir['test/**/*'] +
    %w< citrus.gemspec Rakefile README.md CHANGES >

  s.test_files = s.files.select {|path| path =~ /^test\/.*_test.rb/ }

  s.add_development_dependency('rake')
  s.add_development_dependency('test-unit')

  s.rdoc_options = %w< --line-numbers --inline-source --title Citrus --main Citrus >
  s.extra_rdoc_files = %w< README.md CHANGES >

  s.homepage = 'http://mjackson.github.io/citrus'
  s.licenses = ['MIT']
end

