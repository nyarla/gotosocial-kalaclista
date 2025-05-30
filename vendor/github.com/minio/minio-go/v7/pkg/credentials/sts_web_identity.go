/*
 * MinIO Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2019-2022 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package credentials

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// AssumeRoleWithWebIdentityResponse contains the result of successful AssumeRoleWithWebIdentity request.
type AssumeRoleWithWebIdentityResponse struct {
	XMLName          xml.Name          `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleWithWebIdentityResponse" json:"-"`
	Result           WebIdentityResult `xml:"AssumeRoleWithWebIdentityResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

// WebIdentityResult - Contains the response to a successful AssumeRoleWithWebIdentity
// request, including temporary credentials that can be used to make MinIO API requests.
type WebIdentityResult struct {
	AssumedRoleUser AssumedRoleUser `xml:",omitempty"`
	Audience        string          `xml:",omitempty"`
	Credentials     struct {
		AccessKey    string    `xml:"AccessKeyId" json:"accessKey,omitempty"`
		SecretKey    string    `xml:"SecretAccessKey" json:"secretKey,omitempty"`
		Expiration   time.Time `xml:"Expiration" json:"expiration,omitempty"`
		SessionToken string    `xml:"SessionToken" json:"sessionToken,omitempty"`
	} `xml:",omitempty"`
	PackedPolicySize            int    `xml:",omitempty"`
	Provider                    string `xml:",omitempty"`
	SubjectFromWebIdentityToken string `xml:",omitempty"`
}

// WebIdentityToken - web identity token with expiry.
type WebIdentityToken struct {
	Token        string
	AccessToken  string
	RefreshToken string
	Expiry       int
}

// A STSWebIdentity retrieves credentials from MinIO service, and keeps track if
// those credentials are expired.
type STSWebIdentity struct {
	Expiry

	// Optional http Client to use when connecting to MinIO STS service.
	// (overrides default client in CredContext)
	Client *http.Client

	// Exported STS endpoint to fetch STS credentials.
	STSEndpoint string

	// Exported GetWebIDTokenExpiry function which returns ID
	// tokens from IDP. This function should return two values
	// one is ID token which is a self contained ID token (JWT)
	// and second return value is the expiry associated with
	// this token.
	// This is a customer provided function and is mandatory.
	GetWebIDTokenExpiry func() (*WebIdentityToken, error)

	// RoleARN is the Amazon Resource Name (ARN) of the role that the caller is
	// assuming.
	RoleARN string

	// Policy is the policy where the credentials should be limited too.
	Policy string

	// roleSessionName is the identifier for the assumed role session.
	roleSessionName string

	// Optional, used for token revokation
	TokenRevokeType string
}

// NewSTSWebIdentity returns a pointer to a new
// Credentials object wrapping the STSWebIdentity.
func NewSTSWebIdentity(stsEndpoint string, getWebIDTokenExpiry func() (*WebIdentityToken, error), opts ...func(*STSWebIdentity)) (*Credentials, error) {
	if getWebIDTokenExpiry == nil {
		return nil, errors.New("Web ID token and expiry retrieval function should be defined")
	}
	i := &STSWebIdentity{
		STSEndpoint:         stsEndpoint,
		GetWebIDTokenExpiry: getWebIDTokenExpiry,
	}
	for _, o := range opts {
		o(i)
	}
	return New(i), nil
}

// NewKubernetesIdentity returns a pointer to a new
// Credentials object using the Kubernetes service account
func NewKubernetesIdentity(stsEndpoint string, opts ...func(*STSWebIdentity)) (*Credentials, error) {
	return NewSTSWebIdentity(stsEndpoint, func() (*WebIdentityToken, error) {
		token, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err != nil {
			return nil, err
		}

		return &WebIdentityToken{
			Token: string(token),
		}, nil
	}, opts...)
}

// WithPolicy option will enforce that the returned credentials
// will be scoped down to the specified policy
func WithPolicy(policy string) func(*STSWebIdentity) {
	return func(i *STSWebIdentity) {
		i.Policy = policy
	}
}

func getWebIdentityCredentials(clnt *http.Client, endpoint, roleARN, roleSessionName string, policy string,
	getWebIDTokenExpiry func() (*WebIdentityToken, error), tokenRevokeType string,
) (AssumeRoleWithWebIdentityResponse, error) {
	idToken, err := getWebIDTokenExpiry()
	if err != nil {
		return AssumeRoleWithWebIdentityResponse{}, err
	}

	v := url.Values{}
	v.Set("Action", "AssumeRoleWithWebIdentity")
	if len(roleARN) > 0 {
		v.Set("RoleArn", roleARN)

		if len(roleSessionName) == 0 {
			roleSessionName = strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		v.Set("RoleSessionName", roleSessionName)
	}
	v.Set("WebIdentityToken", idToken.Token)
	if idToken.AccessToken != "" {
		// Usually set when server is using extended userInfo endpoint.
		v.Set("WebIdentityAccessToken", idToken.AccessToken)
	}
	if idToken.RefreshToken != "" {
		// Usually set when server is using extended userInfo endpoint.
		v.Set("WebIdentityRefreshToken", idToken.RefreshToken)
	}
	if idToken.Expiry > 0 {
		v.Set("DurationSeconds", fmt.Sprintf("%d", idToken.Expiry))
	}
	if policy != "" {
		v.Set("Policy", policy)
	}
	v.Set("Version", STSVersion)
	if tokenRevokeType != "" {
		v.Set("TokenRevokeType", tokenRevokeType)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return AssumeRoleWithWebIdentityResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return AssumeRoleWithWebIdentityResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := clnt.Do(req)
	if err != nil {
		return AssumeRoleWithWebIdentityResponse{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return AssumeRoleWithWebIdentityResponse{}, err
		}
		_, err = xmlDecodeAndBody(bytes.NewReader(buf), &errResp)
		if err != nil {
			var s3Err Error
			if _, err = xmlDecodeAndBody(bytes.NewReader(buf), &s3Err); err != nil {
				return AssumeRoleWithWebIdentityResponse{}, err
			}
			errResp.RequestID = s3Err.RequestID
			errResp.STSError.Code = s3Err.Code
			errResp.STSError.Message = s3Err.Message
		}
		return AssumeRoleWithWebIdentityResponse{}, errResp
	}

	a := AssumeRoleWithWebIdentityResponse{}
	if err = xml.NewDecoder(resp.Body).Decode(&a); err != nil {
		return AssumeRoleWithWebIdentityResponse{}, err
	}

	return a, nil
}

// RetrieveWithCredContext is like Retrieve with optional cred context.
func (m *STSWebIdentity) RetrieveWithCredContext(cc *CredContext) (Value, error) {
	if cc == nil {
		cc = defaultCredContext
	}

	client := m.Client
	if client == nil {
		client = cc.Client
	}
	if client == nil {
		client = defaultCredContext.Client
	}

	stsEndpoint := m.STSEndpoint
	if stsEndpoint == "" {
		stsEndpoint = cc.Endpoint
	}
	if stsEndpoint == "" {
		return Value{}, errors.New("STS endpoint unknown")
	}

	a, err := getWebIdentityCredentials(client, stsEndpoint, m.RoleARN, m.roleSessionName, m.Policy, m.GetWebIDTokenExpiry, m.TokenRevokeType)
	if err != nil {
		return Value{}, err
	}

	// Expiry window is set to 10secs.
	m.SetExpiration(a.Result.Credentials.Expiration, DefaultExpiryWindow)

	return Value{
		AccessKeyID:     a.Result.Credentials.AccessKey,
		SecretAccessKey: a.Result.Credentials.SecretKey,
		SessionToken:    a.Result.Credentials.SessionToken,
		Expiration:      a.Result.Credentials.Expiration,
		SignerType:      SignatureV4,
	}, nil
}

// Retrieve retrieves credentials from the MinIO service.
// Error will be returned if the request fails.
func (m *STSWebIdentity) Retrieve() (Value, error) {
	return m.RetrieveWithCredContext(nil)
}

// Expiration returns the expiration time of the credentials
func (m *STSWebIdentity) Expiration() time.Time {
	return m.expiration
}
