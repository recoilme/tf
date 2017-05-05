package main

import (
	"log"
	"testing"
)

func TestAverage(t *testing.T) {
	groupName := "cook_good"
	group := pubDbGet(groupName)
	//if group == nil {
	log.Println("group1", group.ScreenName)
}

func TestVkWallUpd(t *testing.T) {
	log.Println("vkWallUpd")
	vkWallUpd()
}
