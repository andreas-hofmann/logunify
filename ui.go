package main

type UI interface {
	AddData(data LogEntry)
	Update()
	Run()
}
