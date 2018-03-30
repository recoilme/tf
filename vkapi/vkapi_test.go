package vkapi

import (
	"testing"
)

func TestPosts(t *testing.T) {
	s := WallGet("dnodnaru")
	if len(s) != 20 {
		t.Error("Expected 20, got ", len(s))
	}
}

func TestWallOwn(t *testing.T) {
	s := WallGet(125698500)
	if len(s) != 20 {
		t.Error("WO Expected 20, got ", len(s))
	}
}

func TestGroup(t *testing.T) {
	test := "cook_good"
	groups := GroupsGetById(test)

	if len(groups) != 1 {
		t.Error("Expected 1, got ", len(groups))
	}
	if groups[0].ScreenName != test {
		t.Error("Expected "+test, groups[0].ScreenName)
	}
}
