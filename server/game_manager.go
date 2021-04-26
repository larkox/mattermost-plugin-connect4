package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/larkox/mattermost-plugin-connect4/server/connect4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	whiteTag   = "white"
	blackTag   = "black"
	channelTag = "channel"
	postTag    = "post"
)

type GameManager struct {
	api              plugin.API
	botID            string
	grantAchievement func(name string, userID string)
	getAttachmentURL func() string
	getImageURL      func(id string) string
}

func NewGameManager(
	api plugin.API,
	botID string,
	grantAchievement func(name string, userID string),
	getAttachmentURL func() string,
	getImageURL func(id string) string,
) GameManager {
	return GameManager{
		api:              api,
		botID:            botID,
		grantAchievement: grantAchievement,
		getAttachmentURL: getAttachmentURL,
		getImageURL:      getImageURL,
	}
}

func (gm *GameManager) CreateGame(playerA, playerB string) error {
	c, appErr := gm.api.GetDirectChannel(playerA, playerB)
	if appErr != nil {
		return appErr
	}

	originalGame := gm.getGame(c.Id)
	if originalGame != nil {
		if originalGame.Outcome() == connect4.OutcomeNoOutcome {
			return errors.New("still an active game")
		}
	}

	game := connect4.NewGame(playerA, playerB, c.Id)

	post, appErr := gm.api.CreatePost(gm.gameToPost(game))
	if appErr != nil {
		return appErr
	}

	game.SetPostID(post.Id)
	gm.saveGame(game)
	return nil
}

func (gm *GameManager) Move(id, player string, movement int) error {
	game := gm.getGame(id)
	if game == nil {
		return errors.New("no game started")
	}

	turn := game.GetTurnPlayer()
	if player != turn {
		return errors.New("it is not your turn")
	}

	err := game.Move(movement)
	if err != nil {
		return err
	}

	gm.saveGame(game)
	return nil
}

func (gm *GameManager) Resign(id, player string) error {
	game := gm.getGame(id)
	if game == nil {
		return errors.New("no game started")
	}

	_, _, player1, player2 := gm.getGameMetadata(game)

	switch player {
	case player1.Id:
		game.Resign(connect4.Player1)
	case player2.Id:
		game.Resign(connect4.Player2)
	default:
		return errors.New("you are not playing")
	}

	gm.saveGame(game)
	return nil
}

func (gm *GameManager) getGame(id string) connect4.Game {
	b, appErr := gm.api.KVGet(id)
	if appErr != nil {
		return nil
	}

	game, err := connect4.GameFromJSON(b)
	if err != nil {
		return nil
	}

	return game
}

func (gm *GameManager) saveGame(game connect4.Game) {
	id, _, _, _ := gm.getGameMetadata(game)
	_ = gm.api.KVSet(id, game.ToJSON())
}

func (gm *GameManager) gameToPost(game connect4.Game) *model.Post {
	channelID, postID, player1, player2 := gm.getGameMetadata(game)

	post := &model.Post{
		Id:        postID,
		ChannelId: channelID,
		UserId:    gm.botID,
	}

	turnID := game.GetTurnPlayer()
	var turn string
	switch turnID {
	case player1.Id:
		turn = player1.Username
	case player2.Id:
		turn = player2.Username
	}

	attachment := &model.SlackAttachment{
		Title:    "Connect4 game",
		ImageURL: gm.getImageURL(channelID),
		Text:     fmt.Sprintf("Player1: %s\nPlayer2: %s\nTurn: %s", player1.Username, player2.Username, turn),
	}

	switch game.Outcome() {
	case connect4.OutcomeNoOutcome:
		attachment.Actions = []*model.PostAction{
			{
				Type: "button",
				Name: "Move",
				Integration: &model.PostActionIntegration{
					URL: gm.getAttachmentURL() + AttachmentPathMove + "/" + channelID,
				},
			},
			{
				Type: "button",
				Name: "Resign",
				Integration: &model.PostActionIntegration{
					URL: gm.getAttachmentURL() + AttachmentPathResign + "/" + channelID,
				},
			},
		}
	case connect4.OutcomePlayer1Win:
		gm.grantAchievement(AchievementNameWinner, player1.Id)
		attachment.Footer = "Player1 won!"
	case connect4.OutcomePlayer2Win:
		gm.grantAchievement(AchievementNameWinner, player2.Id)
		attachment.Footer = "Player2 won!"
	case connect4.OutcomePlayer1Resign:
		gm.grantAchievement(AchievementNameWinner, player2.Id)
		attachment.Footer = "Player2 won because player 1 resigned!"
	case connect4.OutcomePlayer2Resign:
		gm.grantAchievement(AchievementNameWinner, player1.Id)
		attachment.Footer = "Player1 won because Player2 resigned!"
	case connect4.OutcomeDraw:
		attachment.Footer = "Draw!"
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	return post
}

func (gm *GameManager) getGameMetadata(game connect4.Game) (string, string, *model.User, *model.User) {
	id, postID, player1ID, player2ID := game.GetMetadata()

	player1, appErr := gm.api.GetUser(player1ID)
	if appErr != nil {
		return "", "", nil, nil
	}
	player2, appErr := gm.api.GetUser(player2ID)
	if appErr != nil {
		return "", "", nil, nil
	}

	return id, postID, player1, player2
}

func (gm *GameManager) GetGamePost(id string) *model.Post {
	g := gm.getGame(id)
	if g == nil {
		return nil
	}

	return gm.gameToPost(g)
}

func (gm *GameManager) CanMove(id, player string) bool {
	g := gm.getGame(id)
	if g == nil {
		return false
	}

	return g.GetTurnPlayer() == player
}

func (gm *GameManager) IsPlayingGame(id, player string) bool {
	g := gm.getGame(id)
	if g == nil {
		return false
	}

	_, _, player1, player2 := gm.getGameMetadata(g)

	return player1.Id == player || player2.Id == player
}

func (gm *GameManager) PrintImage(w http.ResponseWriter, id string) {
	g := gm.getGame(id)
	if g == nil {
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	g.EncodeBoard(w)
}

func (gm *GameManager) ValidMovements(id string) []int {
	g := gm.getGame(id)
	if g == nil {
		return nil
	}

	return g.ValidMovements()
}
