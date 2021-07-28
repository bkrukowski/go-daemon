package provider

// NewDefault creates default config provider
func NewDefault(fileFinder func() string) *ConfigProvider {
	return New([]configProcessor{
		newYamlReader(fileFinder),
		newTemplateReplacer(),
		newVarsReplacer(),
		newEnvAppender(),
		newTemplatesCompiler(),
	})
}
