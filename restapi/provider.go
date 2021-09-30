package restapi

import (
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*Provider implements the REST API provider*/
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"uri": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("REST_API_URI", nil),
				Description: "URI of the REST API endpoint. This serves as the base of all requests.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("REST_API_INSECURE", nil),
				Description: "When using https, this disables TLS verification of the host.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("REST_API_TIMEOUT", 0),
				Description: "When set, will cause requests taking longer than this time (in seconds) to be aborted.",
			},
			"rate_limit": {
				Type:        schema.TypeFloat,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("REST_API_RATE_LIMIT", math.MaxFloat64),
				Description: "Set this to limit the number of requests per second made to the API.",
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("REST_API_DEBUG", nil),
				Description: "Enabling this will cause lots of debug information to be printed to STDOUT by the API client.",
			},
			"oauth_client_credentials": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration for oauth client credential flow",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oauth_client_id": {
							Type:        schema.TypeString,
							Description: "client id",
							Required:    true,
						},
						"oauth_client_secret": {
							Type:        schema.TypeString,
							Description: "client secret",
							Required:    true,
						},
						"oauth_token_endpoint": {
							Type:        schema.TypeString,
							Description: "oauth token endpoint",
							Required:    true,
						},
						"oauth_scopes": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Description: "scopes",
						},
					},
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			/* Could only get terraform to recognize this resource if
			         the name began with the provider's name and had at least
				 one underscore. This is not documented anywhere I could find */
			"radius_profile": resourceProfile(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
		ConfigureFunc:  configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	opt := &apiClientOpt{
		uri:       d.Get("uri").(string),
		insecure:  d.Get("insecure").(bool),
		timeout:   d.Get("timeout").(int),
		rateLimit: d.Get("rate_limit").(float64),
		debug:     d.Get("debug").(bool),
	}

	if v, ok := d.GetOk("oauth_client_credentials"); ok {
		oauthConfig := v.([]interface{})[0].(map[string]interface{})

		opt.oauthClientID = oauthConfig["oauth_client_id"].(string)
		opt.oauthClientSecret = oauthConfig["oauth_client_secret"].(string)
		opt.oauthTokenURL = oauthConfig["oauth_token_endpoint"].(string)
		opt.oauthScopes = expandStringSet(oauthConfig["oauth_scopes"].([]interface{}))

	}

	client, err := NewAPIClient(opt)

	return client, err
}
