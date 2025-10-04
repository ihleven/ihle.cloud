package gitauth

import (
	"time"

	"bitbucket.org/hotelplan/webcc-content/cms/content"
	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/auth"
	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

func (c *authProvider) ListGroups() []*auth.Group {
	var groups []*auth.Group
	for _, group := range c.Groups {
		groups = append(groups, group)
	}
	return groups
}

func (c *authProvider) GetGroup(grp string) (*auth.Group, error) {
	group, ok := c.Groups[grp]
	if !ok {
		return nil, errors.NewWithCode(404, "group not found")
	}
	return group, nil
}

func (c *authProvider) CreateGroup(data auth.Group, a *auth.Account) (*auth.Group, error) {

	now := time.Now()
	entry := content.Entry{
		Meta: content.Meta{
			ID:    "group:" + data.Name,
			Space: "internal",
			Repo:  "internal",
			Path:  "groups/" + data.Name + ".json",
			Type:  "Group",
			MIME:  "application/json",
			// Suffix:     ".json",
			Version:    "pub",
			Status:     "active",
			Collection: "groups",
			Access: content.Access{
				Owner:     a.ID,
				Created:   now,
				Modified:  now,
				Published: now,
			},
		},
		Content: &data,
	}
	err := c.commit(a, "create group "+data.Name, "internal", map[string]*content.Entry{entry.Path: &entry})
	if err != nil {
		return nil, err
	}

	c.Groups[data.Name] = &data

	return entry.Content.(*auth.Group), nil
}

func (c *authProvider) UpdateGroup(data auth.Group, a *auth.Account) (*content.Entry, error) {

	_, ok := c.Groups[data.Name]
	if !ok {
		return nil, errors.NewWithCode(404, "group not found: %s", data.Name)
	}
	bytes, err := c.repo.GetBytes("groups/"+data.Name+".json", false)
	if err != nil {
		return nil, err
	}
	entry, err := content.ParseEntry(nil, bytes, "application/json")
	if err != nil {
		return nil, err
	}
	entry.Content = &data

	err = c.commit(a, "upate group "+data.Name, "internal", map[string]*content.Entry{entry.Path: entry})
	if err != nil {
		return nil, err
	}
	c.Groups[data.Name] = &data
	return entry, nil // entry.Content.(*Account), nil
}
func (c *authProvider) DeleteGroup(id string, auth *auth.Account) error {

	delete(c.Groups, id)
	if _, ok := c.Groups[id]; ok {
		return errors.New("delete failed")
	}
	return nil
}
