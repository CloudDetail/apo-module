package elasticsearch

type ElasticsearchConfig struct {
	EsIndexSuffix string   `mapstructure:"es_index_suffix"`
	URLS          []string `mapstructure:"urls"`
	Username      string   `mapstructure:"username"`
	Password      string   `mapstructure:"password"`
}
