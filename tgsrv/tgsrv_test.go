package main

import (
	"log"
	"testing"
)

func TestAverage(t *testing.T) {
	groupName := "myakotkapub"
	group := pubDbGet(groupName)
	//if group == nil {
	log.Println("group1", group)
}

func TestVkWallUpd(t *testing.T) {
	log.Println("vkWallUpd")
	//vkapi.vkWallUpd()
}
