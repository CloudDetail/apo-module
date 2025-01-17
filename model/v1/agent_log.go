package model

type CameraAgentLog struct {
	Contents LogContent `json:"contents"`
	//Tags...
}

type LogContent struct {
	Content string `json:"content"`
	//Sources
}
