package moebotApi

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	botDb "github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/gorilla/mux"

	"github.com/camd67/moebot/moebot_www/db"
)

type MoebotEventsApi struct {
	moebotDatabaseEndpoint
	moebotDiscordEndpoint
}

func NewEventsApi(mdb *db.MoeWebDb, session *discordgo.Session) *MoebotEventsApi {
	return &MoebotEventsApi{moebotDatabaseEndpoint{mdb}, moebotDiscordEndpoint{session}}
}

func (a *MoebotEventsApi) Binds() []bind {
	return []bind{
		bind{Path: "/{guildUID}/events", PermissionLevel: botDb.PermMod, Methods: []string{"GET"}, Handler: a.GetEvents},
		bind{Path: "/{guildUID}/events", PermissionLevel: botDb.PermMod, Methods: []string{"POST"}, Handler: a.AddEvent},
		bind{Path: "/{guildUID}/events/{eventID}", PermissionLevel: botDb.PermMod, Methods: []string{"PUT"}, Handler: a.UpdateEvent},
		bind{Path: "/{guildUID}/events/{eventID}", PermissionLevel: botDb.PermMod, Methods: []string{"DELETE"}, Handler: a.DeleteEvent},
	}
}

func (a *MoebotEventsApi) GetEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	from, err := time.Parse(time.RFC3339, vars["From"])
	if err != nil {
		log.Println("Cannot parse FROM date for request", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorResponse("Invalid FROM date"))
		return
	}
	to, err := time.Parse(time.RFC3339, vars["To"])
	if err != nil {
		log.Println("Cannot parse TO date for request", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorResponse("Invalid TO date"))
		return
	}

	events, err := a.moeWebDb.GuildEvents.SelectEvents(from, to, vars["GuildUID"])
	response, _ := json.Marshal(events)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *MoebotEventsApi) AddEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	event := &db.GuildEvent{}
	if err := json.NewDecoder(r.Body).Decode(event); err != nil {
		log.Println("Cannot parse request body", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorResponse("Cannot parse the request"))
		return
	}

	event.GuildUID = vars["GuildUID"]
	event, err := a.moeWebDb.GuildEvents.InsertEvent(event)
	if err != nil {
		log.Println("Cannot insert new event", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errorResponse("Cannot insert new event"))
		return
	}

	response, _ := json.Marshal(event)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *MoebotEventsApi) UpdateEvent(w http.ResponseWriter, r *http.Request) {

}

func (a *MoebotEventsApi) DeleteEvent(w http.ResponseWriter, r *http.Request) {

}
