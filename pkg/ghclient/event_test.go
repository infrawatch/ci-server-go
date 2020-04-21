package ghclient

import (
	"testing"

	"github.com/infrawatch/ci-server-go/pkg/assert"
)

func TestHandle(t *testing.T) {
	t.Run("push event handle", func(t *testing.T) {
		testRepoJSON := []byte(`{"repository":{"id":186853002,"node_id":"MDEwOlJlcG9zaXRvcnkxODY4NTMwMDI=","name":"Hello-World","full_name":"Codertocat/Hello-World","private":false,"owner":{"name":"Codertocat","email":"21031067+Codertocat@users.noreply.github.com","login":"Codertocat","id":21031067,"node_id":"MDQ6VXNlcjIxMDMxMDY3","avatar_url":"https://avatars1.githubusercontent.com/u/21031067?v=4","gravatar_id":"","url":"https://api.github.com/users/Codertocat","html_url":"https://github.com/Codertocat","followers_url":"https://api.github.com/users/Codertocat/followers","following_url":"https://api.github.com/users/Codertocat/following{/other_user}","gists_url":"https://api.github.com/users/Codertocat/gists{/gist_id}","starred_url":"https://api.github.com/users/Codertocat/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/Codertocat/subscriptions","organizations_url":"https://api.github.com/users/Codertocat/orgs","repos_url":"https://api.github.com/users/Codertocat/repos","events_url":"https://api.github.com/users/Codertocat/events{/privacy}","received_events_url":"https://api.github.com/users/Codertocat/received_events","type":"User","site_admin":false}}}`)

		pushEvent := &Push{}

		repoUnderTest, err := pushEvent.Handle(testRepoJSON)
		if err != nil {
			assert.Ok(t, err)
		}

		expRepo := &Repository{
			Name: "Hello-World",
			Fork: false,
			Owner: struct {
				Login string `json:"login"`
			}{
				Login: "Codertocat",
			},
			refs: make(map[string]*Reference),
		}
		assert.Equals(t, expRepo, repoUnderTest)
	})
}