/*
Copyright 2022-2025 The nagare media authors

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

package v1alpha1

import (
	"errors"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	HeadersQueryArgKey = "nme-headers"
)

const (
	HTTPBasicAuthUsernameKey = "username"
	HTTPBasicAuthPasswordKey = "password"

	HTTPTokenAuthTokenKey = "token"

	HTTPTokenAuthDefaultHeader            = "Authorization"
	HTTPTokenAuthDefaultHeaderValuePrefix = "Bearer"
)

// Specification of a media location.
type MediaLocationSpec struct {
	MediaLocationConfig `json:",inline"`

	// Timeout of the connection.
	// TODO(mtneug): is this a universal option?
	// +optional
	// Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// Configuration of a media location.
// Exactly one of these must be set.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type MediaLocationConfig struct {
	// Configures an HTTP media location.
	// This media location can be used between NBMP tasks that use the "step" execution mode.
	// +optional
	HTTP *HTTPMediaLocation `json:"http,omitempty"`

	// Configures an S3 media location.
	// This media location can be used between NBMP tasks that use the "step" execution mode.
	// +optional
	S3 *S3MediaLocation `json:"s3,omitempty"`

	// Configures an Opencast media location.
	// This media location can be used between NBMP tasks that use the "step" execution mode.
	// +optional
	Opencast *OpencastMediaLocation `json:"opencast,omitempty"`

	// Configures an RTMP media location.
	// This media location can be used between NBMP tasks that use the "streaming" execution mode.
	// +optional
	RTMP *RTMPMediaLocation `json:"rtmp,omitempty"`

	// Configures an RTSP media location.
	// This media location can be used between NBMP tasks that use the "streaming" execution mode.
	// +optional
	RTSP *RTSPMediaLocation `json:"rtsp,omitempty"`

	// Configures a RIST media location.
	// This media location can be used between NBMP tasks that use the "streaming" execution mode.
	// +optional
	RIST *RISTMediaLocation `json:"rist,omitempty"`

	// TODO(mtneug): potential additional media locations:
	//               SRT, RTP, FTP, SFTP, Swift, Google Cloud Storage, Azure Blob Storage
}

func (ml *MediaLocationConfig) URL() (*url.URL, error) {
	switch {
	case ml.HTTP != nil:
		return ml.HTTP.URL()
	case ml.S3 != nil:
		return ml.S3.URL()
	case ml.Opencast != nil:
		return ml.Opencast.URL()
	case ml.RTMP != nil:
		return ml.RTMP.URL()
	case ml.RTSP != nil:
		return ml.RTSP.URL()
	case ml.RIST != nil:
		return ml.RIST.URL()
	}
	return nil, errors.New("invalid MediaLocation: no configuration specified")
}

func (ml *MediaLocationConfig) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration of an HTTP media location.
type HTTPMediaLocation struct {
	// HTTP base URL. Media referencing this location are relative to this URL.
	// +kubebuilder:validation:Pattern="^(http|https)://.*$"
	BaseURL string `json:"baseURL"`

	// List of HTTP headers that should be send with HTTP requests.
	// Note that it is up the the function implementation to honor these headers.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	Headers []Header `json:"headers,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// List of HTTP query arguments that should be send with HTTP requests.
	// Note that it is up the the function implementation to honor these query arguments.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	QueryArgs []QueryArg `json:"queryArgs,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// HTTP authentication configuration.
	// +optional
	Auth *HTTPAuthConfig `json:"auth,omitempty"`
}

func (ml *HTTPMediaLocation) URL() (*url.URL, error) {
	u, err := url.Parse(ml.BaseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for _, arg := range ml.QueryArgs {
		val := ""
		if arg.Value != nil {
			val = *arg.Value
		}
		q.Add(arg.Name, val)
	}

	for _, h := range ml.Headers {
		val := ""
		if h.Value != nil {
			val = *h.Value
		}
		q.Add(HeadersQueryArgKey, url.QueryEscape(h.Name+"="+val))
	}

	switch {
	case ml.Auth.Basic != nil:
		user := ml.Auth.Basic.SecretRef.Data[HTTPBasicAuthUsernameKey]
		pass := ml.Auth.Basic.SecretRef.Data[HTTPBasicAuthPasswordKey]
		if user != nil {
			if pass != nil {
				u.User = url.UserPassword(string(user), string(pass))
			} else {
				u.User = url.User(string(user))
			}
		}

	case ml.Auth.Token != nil:
		token := ml.Auth.Token.SecretRef.Data[HTTPTokenAuthTokenKey]
		if token != nil {
			header := HTTPTokenAuthDefaultHeader
			if ml.Auth.Token.HeaderName != nil {
				header = *ml.Auth.Token.HeaderName
			}

			prefix := HTTPTokenAuthDefaultHeaderValuePrefix
			if ml.Auth.Token.HeaderValuePrefix != nil {
				prefix = *ml.Auth.Token.HeaderValuePrefix
			}

			val := string(token)
			if prefix != "" {
				val = prefix + " " + val
			}

			q.Set(HeadersQueryArgKey, url.QueryEscape(header+"="+val))
		}
	}

	u.RawQuery = q.Encode()

	return u, nil
}

func (ml *HTTPMediaLocation) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration of an HTTP authentication method.
// At most one of these must be set.
// +kubebuilder:validation:MinProperties=0
// +kubebuilder:validation:MaxProperties=1
type HTTPAuthConfig struct {
	// Configures an HTTP basic authentication method.
	// +optional
	Basic *HTTPBasicAuth `json:"basic,omitempty"`

	// Configures an HTTP bearer token authentication method.
	// +optional
	Token *HTTPTokenAuth `json:"token,omitempty"`
}

// Configuration of an HTTP basic authentication method.
type HTTPBasicAuth struct {
	// Reference to a Secret that contains the keys "username" and "password". Only references to Secrets are allowed. A
	// MediaLocation can only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

// Configuration of an HTTP bearer token authentication method.
type HTTPTokenAuth struct {
	// Reference to a Secret that contains the key "token". Only references to Secrets are allowed. A MediaLocation can
	// only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`

	// Name of the HTTP header the token should be passed to. The default is "Authorization".
	// +kubebuilder:default="Authorization"
	// +optional
	HeaderName *string `json:"headerName,omitempty"`

	// Prefix of the HTTP header value before the token. The default is "Bearer".
	// +kubebuilder:default="Bearer"
	// +optional
	HeaderValuePrefix *string `json:"headerValuePrefix"`
}

// Configuration of an S3 media location.
type S3MediaLocation struct {
	// Name of the S3 bucket.
	Bucket string `json:"bucket"`

	// Region of the S3 bucket.
	Region string `json:"region"`

	// S3 authentication configuration.
	Auth S3AuthConfig `json:"auth"`

	// Custom endpoint URL to send S3 requests to.
	// +optional
	EndpointURL string `json:"endpointURL,omitempty"`

	// Whether to use path-style URLs to access S3. By default virtual-hosted–style is used.
	// +kubebuilder:default=false
	// +optional
	UsePathStyle bool `json:"usePathStyle"`
}

func (ml *S3MediaLocation) URL() (*url.URL, error) {
	// TODO: implement
	return nil, errors.New("todo: implement")
}

func (ml *S3MediaLocation) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration of an S3 authentication method.
// Exactly one of these must be set.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type S3AuthConfig struct {
	// Configures an AWS authentication method.
	// +option
	AWS *AWSAuth `json:"aws"`
}

// Configuration of an AWS authentication method.
type AWSAuth struct {
	// Reference to a Secret that contains the keys "accessKeyID" and "secretAccessKey". Only references to Secrets are
	// allowed. A MediaLocation can only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

// Configuration of an Opencast media location. Opencast version 13.x and newer is required.
type OpencastMediaLocation struct {
	// URL to the Opencast service registry. Usually this takes the form of "http://my.tld/services/available.json".
	// +kubebuilder:validation:Pattern="^(http|https)://.*$"
	ServiceRegistryURL string `json:"serviceRegistryURL"`

	// Overwrite specific Opencast API endpoints. These will be used instead of endpoints from the service registry.
	// +optional
	EndpointOverwrites *OpencastEndpointOverwrites `json:"endpointOverwrites,omitempty"`

	// List of additional HTTP headers that should be send with HTTP requests.
	// Note that it is up the the function implementation to honor these headers.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	Headers []Header `json:"headers,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// List of additional HTTP query arguments that should be send with HTTP requests.
	// Note that it is up the the function implementation to honor these query arguments.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	QueryArgs []QueryArg `json:"queryArgs,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// Opencast authentication configuration.
	Auth OpencastAuthConfig `json:"auth"`
}

func (ml *OpencastMediaLocation) URL() (*url.URL, error) {
	// TODO: implement
	return nil, errors.New("todo: implement")
}

func (ml *OpencastMediaLocation) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration for overwriting specific Opencast endpoints. These will be used instead of the endpoints given by the
// Opencast service registry.
type OpencastEndpointOverwrites struct {
	// Overwrite for the External API.
	// +kubebuilder:validation:Pattern="^(http|https)://.*$"
	// +optional
	ExternalAPI *string `json:"externalAPI,omitempty"`
}

// Configuration of an Opencast authentication method.
// Exactly one of these must be set.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type OpencastAuthConfig struct {
	// Configures an HTTP basic authentication method.
	// +optional
	Basic *HTTPBasicAuth `json:"basic,omitempty"`
}

// Configuration of an RTMP media location.
type RTMPMediaLocation struct {
	// RTMP base URL. Media referencing this location are relative to this URL.
	// +kubebuilder:validation:Pattern="^(rtmp|rtmpe|rtmps|rtmpt|rtmpte|rtmpts)://.*$"
	BaseURL string `json:"baseURL"`

	// The RTMP application name. This overwrites application names given through baseURL.
	// +optional
	App *string `json:"app,omitempty"`

	// List of RTMP query arguments that should be send with RTMP requests.
	// Note that it is up the the function implementation to honor these query arguments.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	QueryArgs []QueryArg `json:"queryArgs,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// RTMP authentication configuration.
	// +optional
	Auth *RTMPAuthConfig `json:"auth,omitempty"`
}

func (ml *RTMPMediaLocation) URL() (*url.URL, error) {
	// TODO: implement
	return nil, errors.New("todo: implement")
}

func (ml *RTMPMediaLocation) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration of an RTMP authentication method.
// Multiple methods can be set.
type RTMPAuthConfig struct {
	// Configures an RTMP basic authentication method.
	// +optional
	Basic *RTMPBasicAuth `json:"basic,omitempty"`

	// Configures an RTMP streaming key authentication method. The streaming key will be used as RTMP playpath.
	// +optional
	StreamingKey *RTMPStreamingKeyAuth `json:"streamingKey,omitempty"`
}

// Configuration of an RTMP basic authentication method.
type RTMPBasicAuth struct {
	// Reference to a Secret that contains the keys "username" and "password". Only references to Secrets are allowed. A
	// MediaLocation can only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

type RTMPStreamingKeyAuth struct {
	// Reference to a Secret that contains the key "streamingKey". Only references to Secrets are allowed. A MediaLocation
	// can only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

// Configuration of an RTSP media location.
type RTSPMediaLocation struct {
	// RTSP base URL. Media referencing this location are relative to this URL.
	// +kubebuilder:validation:Pattern="^(rtsp|rtsps|rtspu)://.*$"
	BaseURL string `json:"baseURL"`

	// Forces a specific transport protocol. The default is "auto" which tries detecting the best transport protocol
	// automatically.
	// +kubebuilder:validation:Enum=auto;udp;tcp;udp_multicast;http;https
	// +kubebuilder:default="auto"
	// +optional
	TransportProtocol *string `json:"transportProtocol,omitempty"`

	// List of RTSP query arguments that should be send with RTSP requests.
	// Note that it is up the the function implementation to honor these query arguments.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	QueryArgs []QueryArg `json:"queryArgs,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// RTSP authentication configuration.
	// +optional
	Auth *RTSPAuthConfig `json:"auth,omitempty"`
}

func (ml *RTSPMediaLocation) URL() (*url.URL, error) {
	// TODO: implement
	return nil, errors.New("todo: implement")
}

func (ml *RTSPMediaLocation) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration of an RTSP authentication method.
// At most one of these must be set.
// +kubebuilder:validation:MinProperties=0
// +kubebuilder:validation:MaxProperties=1
type RTSPAuthConfig struct {
	// Configures an RTSP basic authentication method.
	// +optional
	Basic *RTSPBasicAuth `json:"basic,omitempty"`
}

// Configuration of an RTSP basic authentication method.
type RTSPBasicAuth struct {
	// Reference to a Secret that contains the keys "username" and "password". Only references to Secrets are allowed. A
	// MediaLocation can only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

// Configuration of a RIST media location.
type RISTMediaLocation struct {
	// RIST base URL. Media referencing this location are relative to this URL.
	// +kubebuilder:validation:Pattern="^rist://.*$"
	BaseURL string `json:"baseURL"`

	// RIST profile to use. The default is "main".
	// +kubebuilder:validation:Enum=simple;main;advanced
	// +kubebuilder:default="main"
	// +optional
	Profile *string `json:"profile,omitempty"`

	// Sets the buffer size. The maximum duration is 30s.
	// +optional
	BufferSize *metav1.Duration `json:"bufferSize,omitempty"`

	// RIST encryption configuration.
	// +optional
	Encryption *RISTEncryption `json:"encryption,omitempty"`
}

func (ml *RISTMediaLocation) URL() (*url.URL, error) {
	// TODO: implement
	return nil, errors.New("todo: implement")
}

func (ml *RISTMediaLocation) String() string {
	u, _ := ml.URL()
	return u.String()
}

// Configuration of RIST encryption
type RISTEncryption struct {
	// Encryption type.
	// +kubebuilder:validation:Enum=aes-128;aes-256
	Type string `json:"type"`

	// Reference to a Secret that contains the key "secret". Only references to Secrets are allowed. A MediaLocation can
	// only reference Secrets from its own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

// Specifies a header.
type Header struct {
	// Name of the header.
	Name string `json:"name"`

	// Value of the header. This field is required if valueFrom is not specified. If both are specified, value has
	// precedence.
	// +optional
	Value *string `json:"value,omitempty"`

	// TODO: implement
	// Reference to a ConfigMap or Secret that contains the specified key. Only references to ConfigMaps or Secrets are
	// allowed. A MediaLocation can only reference Objects from its own Namespace. This field is required if value is not
	// specified. If both are specified, value has precedence.
	// +optional
	// ValueFrom *meta.ConfigMapOrSecretReference `json:"valueFrom,omitempty"`
}

// Specifies a URL query argument.
type QueryArg struct {
	// Name of the query argument.
	Name string `json:"name"`

	// Value of the query argument. This field is required if valueFrom is not specified. If both are specified, value has
	// precedence.
	// +optional
	Value *string `json:"value,omitempty"`

	// TODO: implement
	// Reference to a ConfigMap or Secret that contains the specified key. Only references to ConfigMaps or Secrets are
	// allowed. A MediaLocation can only reference Objects from its own Namespace. This field is required if value is not
	// specified. If both are specified, value has precedence.
	// +optional
	// ValueFrom *meta.ConfigMapOrSecretReference `json:"valueFrom,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={nagare-all,nme-all}

// MediaLocation is the Schema for the medialocations API
type MediaLocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MediaLocationSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// MediaLocationList contains a list of MediaLocation
type MediaLocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MediaLocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MediaLocation{}, &MediaLocationList{})
}
