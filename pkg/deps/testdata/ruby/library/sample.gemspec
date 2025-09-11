Gem::Specification.new do |spec|
  spec.name          = "sample"
  spec.version       = "0.1.0"
  spec.summary       = "Sample gem"
  spec.description   = "Sample"
  spec.authors       = ["Test"]
  spec.files         = []

  spec.add_runtime_dependency 'rake', '>= 13.0'
  spec.add_development_dependency 'rspec', '~> 3.10'
end
