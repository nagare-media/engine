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

package engineurl

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/nagare-media/models.go/base"
)

const (
	NagareEngineScheme          = "nagare-media-engine"
	TaskURLPathSegment          = "task"
	MediaLocationURLPathSegment = "media"
)

type URL string

func IsEngineURL(u string) bool {
	return strings.HasPrefix(u, NagareEngineScheme+"://")
}

func Parse(u string) (ParsedURL, error) {
	url := URL(u)
	return url.Parse()
}

func (u *URL) Parse() (ParsedURL, error) {
	url, err := url.Parse(string(*u))
	if err != nil {
		return nil, err
	}

	if url.Scheme != NagareEngineScheme {
		return nil, fmt.Errorf("engineurl: unknown scheme: %s", url.Scheme)
	}

	p := strings.TrimLeft(url.EscapedPath(), "/")
	seg := strings.Split(p, "/")

	switch seg[0] {
	case TaskURLPathSegment:
		if len(seg) != 5 {
			return nil, fmt.Errorf("engineurl: invalid task URL path: %s", p)
		}

		return &TaskURL{
			WorkflowID: seg[1],
			TaskID:     seg[2],
			PortName:   seg[3],
			StreamID:   seg[4],
			RawQuery:   url.RawQuery,
		}, nil

	case MediaLocationURLPathSegment:
		if len(seg) != 3 {
			return nil, fmt.Errorf("engineurl: invalid task URL path: %s", p)
		}

		return &MediaLocationURL{
			Name:     seg[1],
			Path:     seg[2],
			RawQuery: url.RawQuery,
		}, nil

	default:
		return nil, fmt.Errorf("engineurl: unknown URL type: %s", seg[0])
	}
}

type ParsedURL interface {
	fmt.Stringer

	URI() base.URI
	URL() *url.URL
}

type TaskURL struct {
	WorkflowID string
	TaskID     string
	PortName   string
	StreamID   string
	RawQuery   string
}

var _ ParsedURL = &TaskURL{}

func (u *TaskURL) String() string {
	var buf strings.Builder

	n := len(
		NagareEngineScheme + "://" + TaskURLPathSegment +
			"/" + /* WorkflowID */
			"/" + /* TaskID     */
			"/" + /* PortName   */
			"/" + /* StreamID   */
			"?", /*  RawQuery   */
	)
	n += len(u.WorkflowID) + len(u.TaskID) + len(u.PortName) + len(u.StreamID) + len(u.RawQuery)
	buf.Grow(n)

	buf.WriteString(NagareEngineScheme)
	buf.WriteString("://")
	buf.WriteString(TaskURLPathSegment)
	buf.WriteString("/")
	buf.WriteString(u.WorkflowID)
	buf.WriteString("/")
	buf.WriteString(u.TaskID)
	buf.WriteString("/")
	buf.WriteString(u.PortName)
	buf.WriteString("/")
	buf.WriteString(u.StreamID)
	buf.WriteString("?")
	buf.WriteString(u.RawQuery)

	return buf.String()
}

func (u *TaskURL) URI() base.URI {
	return base.URI(u.String())
}

func (u *TaskURL) URL() *url.URL {
	return &url.URL{
		Scheme:   NagareEngineScheme,
		Path:     path.Join("/", TaskURLPathSegment, u.WorkflowID, u.TaskID, u.PortName, u.StreamID),
		RawQuery: u.RawQuery,
	}
}

type MediaLocationURL struct {
	Name     string
	Path     string
	RawQuery string
}

var _ ParsedURL = &MediaLocationURL{}

func (u *MediaLocationURL) String() string {
	var buf strings.Builder

	n := len(
		NagareEngineScheme + "://" + MediaLocationURLPathSegment +
			"/" + /* Name		  */
			"/" + /* Path     */
			"?", /*  RawQuery */
	)
	n += len(u.Name) + len(u.Path) + len(u.RawQuery)
	buf.Grow(n)

	buf.WriteString(NagareEngineScheme)
	buf.WriteString("://")
	buf.WriteString(MediaLocationURLPathSegment)
	buf.WriteString("/")
	buf.WriteString(u.Name)
	buf.WriteString("/")
	buf.WriteString(u.Path)
	buf.WriteString("?")
	buf.WriteString(u.RawQuery)

	return buf.String()
}

func (u *MediaLocationURL) URI() base.URI {
	return base.URI(u.String())
}

func (u *MediaLocationURL) URL() *url.URL {
	return &url.URL{
		Scheme:   NagareEngineScheme,
		Path:     path.Join("/", MediaLocationURLPathSegment, u.Name, u.Path),
		RawQuery: u.RawQuery,
	}
}
