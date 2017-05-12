package main

import (
	"log"
	"testing"
)

func TestRssdomains(t *testing.T) {
	domains := rssdomains()
	log.Println("group1", domains)
}

func TestRss2(t *testing.T) {
	//parseRss()
}
