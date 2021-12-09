package places

import (
	"context"
	"net/url"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestNewClient(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want PlaceAPI
	}{
		{
			name: "Create Clinet",
			args: args{
				key: "api-key",
			},
			want: &Client{
				key: "api-key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_TextSearch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", `=~^https://maps.googleapis.com/maps/api/place/textsearch/json`,
		httpmock.NewStringResponder(200, `{
			"html_attributions": [],
			"status":"OK",
			"results" : [{"business_status" : "OPERATIONAL","formatted_address" : "address","geometry" : {"location" : {"lat" : 35.6951141,"lng" : 139.7926941},"viewport" : {"northeast" : {"lat" : 35.69649627989271,"lng" : 139.7940198298927},"southwest" : {"lat" : 35.69379662010727,"lng" : 139.7913201701073}}},"icon" : "icon.jpg","icon_background_color" : "#FF9E67","icon_mask_base_uri" : "https://icon_mask_base_uri","name" : "beer factory","opening_hours" : {"open_now" : false},"photos" : [{"height" : 4160,"html_attributions" : ["html_attribution1"],"photo_reference" : "photo_reference1","width" : 3120}],"place_id" : "place_id_1","plus_code" : {"compound_code" : "MQWV+23 Tokyo","global_code" : "8Q7XMQWV+23"},"price_level" : 3,"rating" : 4.3,"reference" : "reference1","types" : [ "bar", "restaurant", "food"],"user_ratings_total" : 1047}]
	 }`))

	type fields struct {
		key string
	}
	type args struct {
		ctx    context.Context
		query  string
		params OptionalParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PlaceResult
		wantErr bool
	}{
		{
			name: "textsearch success",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				ctx:   context.Background(),
				query: "london beer",
				params: OptionalParams{
					Language: "en",
					Region:   "uk",
				},
			},
			want: &PlaceResult{
				Status:           "OK",
				HtmlAttributions: []string{},
				Results: []Place{
					{
						FormattedAddress:    "address",
						Icon:                "icon.jpg",
						IconBackgroundColor: "#FF9E67",
						IconMaskBaseUri:     "https://icon_mask_base_uri",
						Name:                "beer factory",
						PlaceID:             "place_id_1",
						PlusCode: PlusCode{
							CompoundCode: "MQWV+23 Tokyo",
							GlobalCode:   "8Q7XMQWV+23",
						},
						PriceLevel:       3,
						Rating:           4.3,
						Reference:        "reference1",
						Types:            []string{"bar", "restaurant", "food"},
						UserRatingsTotal: 1047,
						BusinessStatus:   "OPERATIONAL",
						Geometry: Geometry{
							Location: Location{
								Lat: 35.6951141,
								Lng: 139.7926941,
							},
						},
						Photos: []Photo{
							{
								Height: 4160,
								Width:  3120,
								HtmlAttributions: []string{
									"html_attribution1",
								},
								PhotoReference: "photo_reference1",
							},
						},
						OpeningHours: OpeningHours{
							OpenNow: false,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				key: tt.fields.key,
			}
			got, err := c.TextSearch(tt.args.ctx, tt.args.query, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.TextSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.TextSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_createURL(t *testing.T) {
	wantUrl, _ := url.Parse("http://test-url.com/?key=test-api-key")

	type fields struct {
		key string
	}
	type args struct {
		baseURL string
		params  OptionalParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *url.URL
		wantErr bool
	}{
		{
			name: "parse error",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				baseURL: "http://test-url.com/%%",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "create url",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				baseURL: "http://test-url.com/",
			},
			want:    wantUrl,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				key: tt.fields.key,
			}
			got, err := c.createURL(tt.args.baseURL, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.createURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.createURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_createParams(t *testing.T) {
	u, _ := url.Parse("https://maps.googleapis.com/maps/api/place/textsearch/json")

	type fields struct {
		key string
	}
	type args struct {
		u      *url.URL
		params OptionalParams
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "no optional parameters",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				u:      u,
				params: OptionalParams{},
			},
			want: "key=test-api-key",
		},
		{
			name: "set pagetoken",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				u: u,
				params: OptionalParams{
					Pagetoken: "tokentoken",
				},
			},
			want: "key=test-api-key&pagetoken=tokentoken",
		},
		{
			name: "set all params",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				u: u,
				params: OptionalParams{
					Language: "en",
					Region:   "uk",
					Opennow:  true,
					Radius:   "50000",
					Type:     "cafe",
					Maxprice: "4",
					Minprice: "1",
				},
			},
			want: "key=test-api-key&language=en&maxprice=4&minprice=1&opennow=true&radius=50000&region=uk&type=cafe",
		},
		{
			name: "wrong radius, maxprice, minprice(string)",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				u: u,
				params: OptionalParams{
					Radius:   "radius",
					Maxprice: "maxprice",
					Minprice: "minprice",
				},
			},
			want: "key=test-api-key",
		},
		{
			name: "wrong maxprice, minprice(out of range)",
			fields: fields{
				key: "test-api-key",
			},
			args: args{
				u: u,
				params: OptionalParams{
					Maxprice: "5",
					Minprice: "10",
				},
			},
			want: "key=test-api-key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				key: tt.fields.key,
			}
			if got := c.createParams(tt.args.u, tt.args.params); got != tt.want {
				t.Errorf("Client.createParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_request(t *testing.T) {
	type args struct {
		url      string
		response *PlaceResult
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := request(tt.args.url, tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("request() = %v, want %v", got, tt.want)
			}
		})
	}
}
