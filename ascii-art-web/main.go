package main

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type PageData struct {
	Result string
	Text   string
	Banner string
	Error string
}

var templates *template.Template