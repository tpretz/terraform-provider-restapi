package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var radiusAttribute := &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": &schema.Schema{
			Type: schema.String,
			Required: true,
			ValidateFunc: validation.StringIsNotWhiteSpace,
		},
		"value": &schema.Schema{
			Type: schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"op": &schema.Schema{
			Type: schema.TypeString,
			Optional: true,
			Default: ":="
			ValidateFunc: validation.StringInSlice([]string{"=", ":=", "+=", "|=", "|:=", "|+=", "|--"})
		},
		"expand": &schema.Schema{
			Type: schema.TypeBool,
			Optional: true,
			Default: false
		},
		"do_xlat": &schema.Schema{
			Type: schema.TypeBool,
			Optional: true,
			Default: false
		},
		"is_json": &schema.Schema{
			Type: schema.TypeBool,
			Optional: true,
			Default: false
		},
	},
}

func resourceProfile() *schema.Resource {
	// Consider data sensitive if env variables is set to true.

	return &schema.Resource{
		Create: resourceProfileCreate,
		Read:   resourceProfileRead,
		Update: resourceProfileUpdate,
		Delete: resourceProfileDelete,
		Exists: resourceProfileExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Description:  "Profile ID",
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-z_]{3,32}$"), "must align to regex"),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"weight": {
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(1, 100000),
				Default:      100,
				Optional:     true,
			},
			"depends": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-z_]{3,32}$"), "must align to regex"),
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-zA-Z][0-9a-zA-Z\\,\\.\\_\\-\\' ]{1,512}[0-9a-zA-Z\\.]$"), "must align to regex"),
			},
			"parameter_schema": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"reply": {
				Type:         schema.TypeList,
				Optional:     true,
				Elem: radiusAttribute,
			},
			"control": {
				Type:         schema.TypeList,
				Optional:     true,
				Elem: radiusAttribute,
			},
		}, /* End schema */
	}
}

func buildProfileObject(d *schema.ResourceData, api *APIClient) (interface{}, error) {

}

func resourceProfileCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*APIClient)

	itm, err := buildProfileObject(d, api)

	if err != nil {
		return err
	}

	// do add

	return resourceProfileRead(d, m)
}


	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Create routine called. Object built:\n%s\n", obj.toString())

	err = obj.createObject()
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		d.SetId(obj.id)
		setResourceState(obj, d)
		/* Only set during create for APIs that don't return sensitive data on subsequent retrieval */
		d.Set("create_response", obj.apiResponse)
	}
	return err
}

func resourceProfileRead(d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		if strings.Contains(err.Error(), "error parsing data provided") {
			log.Printf("resource_api_object.go: WARNING! The data passed from Terraform's state is invalid! %v", err)
			log.Printf("resource_api_object.go: Continuing with partially constructed object...")
		} else {
			return err
		}
	}
	log.Printf("resource_api_object.go: Read routine called. Object built:\n%s\n", obj.toString())

	err = obj.readObject()
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		log.Printf("resource_api_object.go: Read resource. Returned id is '%s'\n", obj.id)
		d.SetId(obj.id)
		setResourceState(obj, d)
	}
	return err
}

func resourceProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return err
	}

	/* If copy_keys is not empty, we have to grab the latest
	   data so we can copy anything needed before the update */
	client := meta.(*APIClient)
	if len(client.copyKeys) > 0 {
		err = obj.readObject()
		if err != nil {
			return err
		}
	}

	log.Printf("resource_api_object.go: Update routine called. Object built:\n%s\n", obj.toString())

	err = obj.updateObject()
	if err == nil {
		setResourceState(obj, d)
	}
	return err
}

func resourceProfileDelete(d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Delete routine called. Object built:\n%s\n", obj.toString())

	err = obj.deleteObject()
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			/* 404 means it doesn't exist. Call that good enough */
			err = nil
		}
	}
	return err
}

func resourceProfileExists(d *schema.ResourceData, meta interface{}) (exists bool, err error) {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		if strings.Contains(err.Error(), "error parsing data provided") {
			log.Printf("resource_api_object.go: WARNING! The data passed from Terraform's state is invalid! %v", err)
			log.Printf("resource_api_object.go: Continuing with partially constructed object...")
		} else {
			return exists, err
		}
	}
	log.Printf("resource_api_object.go: Exists routine called. Object built: %s\n", obj.toString())

	/* Assume all errors indicate the object just doesn't exist.
	This may not be a good assumption... */
	err = obj.readObject()
	if err == nil {
		exists = true
	}
	return exists, err
}

/* Simple helper routine to build an api_object struct
   for the various calls terraform will use. Unfortunately,
   terraform cannot just reuse objects, so each CRUD operation
   results in a new object created */
func makeAPIObject(d *schema.ResourceData, meta interface{}) (*APIObject, error) {
	opts, err := buildAPIObjectOpts(d)
	if err != nil {
		return nil, err
	}

	caller := "unknown"
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		parts := strings.Split(details.Name(), ".")
		caller = parts[len(parts)-1]
	}
	log.Printf("resource_rest_api.go: Constructing new APIObject in makeAPIObject (called by %s)", caller)

	obj, err := NewAPIObject(meta.(*APIClient), opts)

	return obj, err
}

func buildAPIObjectOpts(d *schema.ResourceData) (*apiObjectOpts, error) {
	opts := &apiObjectOpts{
		path: d.Get("path").(string),
	}

	/* Allow user to override provider-level id_attribute */
	if v, ok := d.GetOk("id_attribute"); ok {
		opts.idAttribute = v.(string)
	}

	/* Allow user to specify the ID manually */
	if v, ok := d.GetOk("object_id"); ok {
		opts.id = v.(string)
	} else {
		/* If not specified, see if terraform has an ID */
		opts.id = d.Id()
	}

	log.Printf("resource_rest_api.go: buildAPIObjectOpts routine called for id '%s'\n", opts.id)

	if v, ok := d.GetOk("create_path"); ok {
		opts.postPath = v.(string)
	}
	if v, ok := d.GetOk("read_path"); ok {
		opts.getPath = v.(string)
	}
	if v, ok := d.GetOk("update_path"); ok {
		opts.putPath = v.(string)
	}
	if v, ok := d.GetOk("create_method"); ok {
		opts.createMethod = v.(string)
	}
	if v, ok := d.GetOk("read_method"); ok {
		opts.readMethod = v.(string)
	}
	if v, ok := d.GetOk("update_method"); ok {
		opts.updateMethod = v.(string)
	}
	if v, ok := d.GetOk("destroy_method"); ok {
		opts.destroyMethod = v.(string)
	}
	if v, ok := d.GetOk("destroy_path"); ok {
		opts.deletePath = v.(string)
	}
	if v, ok := d.GetOk("query_string"); ok {
		opts.queryString = v.(string)
	}

	readSearch := expandReadSearch(d.Get("read_search").(map[string]interface{}))
	opts.readSearch = readSearch

	opts.data = d.Get("data").(string)
	opts.debug = d.Get("debug").(bool)

	return opts, nil
}

func expandReadSearch(v map[string]interface{}) (readSearch map[string]string) {
	readSearch = make(map[string]string)
	for key, val := range v {
		readSearch[key] = val.(string)
	}

	return
}
