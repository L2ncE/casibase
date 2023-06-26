package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/casbin/casibase/object"
	"github.com/casbin/casibase/util"
	"github.com/casbin/casibase/video"
)

func (c *ApiController) GetGlobalVideos() {
	c.ResponseOk(object.GetGlobalVideos())
}

func (c *ApiController) GetVideos() {
	owner := c.Input().Get("owner")

	c.ResponseOk(object.GetVideos(owner))
}

func (c *ApiController) GetVideo() {
	id := c.Input().Get("id")

	video := object.GetVideo(id)
	video.Populate()

	c.ResponseOk(video)
}

func (c *ApiController) UpdateVideo() {
	id := c.Input().Get("id")

	var video object.Video
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &video)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.UpdateVideo(id, &video))
}

func (c *ApiController) AddVideo() {
	var video object.Video
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &video)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.AddVideo(&video))
}

func (c *ApiController) DeleteVideo() {
	var video object.Video
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &video)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.DeleteVideo(&video))
}

func startCoverUrlJob(owner string, name string, videoId string) {
	go func(owner string, name string, videoId string) {
		for i := 0; i < 20; i++ {
			coverUrl := video.GetVideoCoverUrl(videoId)
			if coverUrl != "" {
				video := object.GetVideo(util.GetIdFromOwnerAndName(owner, name))
				if video.CoverUrl != "" {
					break
				}

				video.CoverUrl = coverUrl
				object.UpdateVideo(util.GetIdFromOwnerAndName(owner, name), video)
				break
			}

			time.Sleep(time.Second * 5)
		}
	}(owner, name, videoId)
}

func (c *ApiController) UploadVideo() {
	owner := c.GetSessionUsername()

	file, header, err := c.GetFile("file")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	defer file.Close()

	filename := header.Filename
	fileId := util.RemoveExt(filename)

	fileBuffer := bytes.NewBuffer(nil)
	if _, err = io.Copy(fileBuffer, file); err != nil {
		c.ResponseError(err.Error())
		return
	}

	fileType := "unknown"
	contentType := header.Header.Get("Content-Type")
	fileType, _ = util.GetOwnerAndNameFromId(contentType)

	if fileType != "video" {
		c.ResponseError(fmt.Sprintf("contentType: %s is not video", contentType))
		return
	}

	videoId := video.UploadVideo(fileId, filename, fileBuffer)
	if videoId != "" {
		startCoverUrlJob(owner, fileId, videoId)

		video := &object.Video{
			Owner:       owner,
			Name:        fileId,
			CreatedTime: util.GetCurrentTime(),
			DisplayName: fileId,
			VideoId:     videoId,
			Labels:      []*object.Label{},
			DataUrls:    []string{},
		}
		object.AddVideo(video)
		c.ResponseOk(fileId)
	} else {
		c.ResponseError("videoId is empty")
	}
}
