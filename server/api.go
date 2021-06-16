package main

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/larkox/mattermost-plugin-connect4/server/connect4"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) initializeAPI() {
	p.router = mux.NewRouter()

	dialogRouter := p.router.PathPrefix(DialogPath).Subrouter()
	dialogRouter.HandleFunc(DialogPathMove+"/{id}", p.handleMovement).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathResign+"/{id}", p.handleResignation).Methods(http.MethodPost)

	attachmentRouter := p.router.PathPrefix(AttachmentPath).Subrouter()
	attachmentRouter.HandleFunc(AttachmentPathMove+"/{id}", p.handleMove).Methods(http.MethodPost)
	attachmentRouter.HandleFunc(AttachmentPathResign+"/{id}", p.handleResign)

	p.router.HandleFunc(ImagePath+"/{id}.svg", p.handleImage).Methods(http.MethodGet)
	p.router.HandleFunc("/test", p.handleTestGame).Methods(http.MethodGet)
}

func (p *Plugin) attachmentError(w http.ResponseWriter, errText string) {
	resp := model.PostActionIntegrationResponse{
		EphemeralText: "Error: " + errText,
	}
	w.Write(resp.ToJson())
}

func (p *Plugin) handleMove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		p.attachmentError(w, "Error: Not authorized")
		return
	}

	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		p.attachmentError(w, "Error: invalid request")
		return
	}

	if !p.gameManager.CanMove(gameID, userID) {
		p.attachmentError(w, "Cannot move.")
		return
	}

	options := []*model.PostActionOptions{}
	for _, m := range p.gameManager.ValidMovements(gameID) {
		options = append(options, &model.PostActionOptions{
			Text:  strconv.Itoa(m),
			Value: strconv.Itoa(m),
		})
	}

	appErr := p.API.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: request.TriggerId,
		URL:       p.getDialogURL() + DialogPathMove + "/" + gameID,
		Dialog: model.Dialog{
			Title: "Make your move",
			IntroductionText: "Select the column where you want to add your next piece.\n\n" +
				"![board](" + p.getImageURL(gameID) + ")",
			SubmitLabel: "Move",
			Elements: []model.DialogElement{
				{
					DisplayName: "Movement",
					Name:        "movement",
					Type:        "select",
					Options:     options,
				},
			},
		},
	})

	if appErr != nil {
		p.API.LogDebug("error opening move", "error", appErr.Error())
	}

	_, _ = w.Write((&model.PostActionIntegrationResponse{}).ToJson())
}

func (p *Plugin) handleResign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		p.attachmentError(w, "Error: Not authorized")
		return
	}

	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		p.attachmentError(w, "Error: invalid request")
		return
	}

	if !p.gameManager.IsPlayingGame(gameID, userID) {
		p.attachmentError(w, "Error: you are not playing this game")
		return
	}

	appErr := p.API.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: request.TriggerId,
		URL:       p.getDialogURL() + DialogPathResign + "/" + gameID,
		Dialog: model.Dialog{
			Title:            "Resign this game?",
			IntroductionText: "Are you sure you want to resign this game?",
			SubmitLabel:      "Resign",
		},
	})

	if appErr != nil {
		p.attachmentError(w, "Error: could not open the interactive dialog, "+appErr.Error())
		return
	}

	_, _ = w.Write((&model.PostActionIntegrationResponse{}).ToJson())
}

func (p *Plugin) handleMovement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		interactiveDialogError(w, "Not authorized", nil)
		return
	}

	request := model.SubmitDialogRequestFromJson(r.Body)
	if request == nil {
		interactiveDialogError(w, "invalid request", nil)
		return
	}

	movementStr, ok := request.Submission["movement"].(string)
	if !ok {
		interactiveDialogError(w, "Invalid field", map[string]string{"movement": "Could not recognize movement."})
		return
	}
	movement, err := strconv.Atoi(movementStr)
	if err != nil {
		interactiveDialogError(w, "Invalid field", map[string]string{"movement": "Could not recognize movement."})
	}

	post, err := p.gameManager.Move(gameID, userID, movement)
	if err != nil {
		interactiveDialogError(w, err.Error(), nil)
		return
	}

	_, _ = p.API.UpdatePost(post)

	_, _ = w.Write((&model.SubmitDialogResponse{}).ToJson())
}

func (p *Plugin) handleResignation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		interactiveDialogError(w, "Not authorized", nil)
		return
	}

	post, err := p.gameManager.Resign(gameID, userID)
	if err != nil {
		interactiveDialogError(w, err.Error(), nil)
		return
	}

	_, _ = p.API.UpdatePost(post)

	_, _ = w.Write((&model.SubmitDialogResponse{}).ToJson())
}

func interactiveDialogError(w http.ResponseWriter, message string, errors map[string]string) {
	resp := model.SubmitDialogResponse{
		Error:  "Error: " + message,
		Errors: errors,
	}

	_, _ = w.Write(resp.ToJson())
}

func (p *Plugin) handleImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		p.attachmentError(w, "Error: Not authorized")
		return
	}

	p.gameManager.PrintImage(w, gameID)
}

// Credit to: https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
var errInvalidFormat = errors.New("invalid format")

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat
	}
	return
}

func (p *Plugin) handleTestGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	g := connect4.NewGame(p.BotUserID, p.BotUserID, "")
	rand.Seed(time.Now().UnixNano())
	player, waiting := 2, 1
	for i := 0; i < 1000; i++ {
		player, waiting = waiting, player
		if g.Outcome() != connect4.OutcomeNoOutcome {
			break
		}

		m := g.ValidMovements()
		if len(m) == 0 {
			fmt.Printf("Should not happen")
			break
		}
		j := rand.Intn(len(m))
		g.Move(m[j])
	}
	g.EncodeBoard(w)
}

func (p *Plugin) getPluginURL() string {
	urlP := p.API.GetConfig().ServiceSettings.SiteURL
	url := "/"
	if urlP != nil {
		url = *urlP
	}
	if url[len(url)-1] == '/' {
		url = url[0 : len(url)-1]
	}
	return url + "/plugins/" + manifest.Id
}

func (p *Plugin) getDialogURL() string {
	return p.getPluginURL() + DialogPath
}

func (p *Plugin) getAttachmentURL() string {
	return p.getPluginURL() + AttachmentPath
}

func (p *Plugin) getImageURL(id string) string {
	return fmt.Sprintf(
		"%s%s/%s.svg?ts=%s",
		p.getPluginURL(),
		ImagePath,
		id,
		time.Now().Format("2006-01-02T15:04:05Z07:00"),
	)
}
