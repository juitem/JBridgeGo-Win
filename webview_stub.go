//go:build !darwin

package main

func platformSetup()              {}
func platformNavigate(url string) {}
func platformGoBack()             {}
func platformGoForward()          {}
func platformReload()             {}
func platformIsMac() bool         { return false }
