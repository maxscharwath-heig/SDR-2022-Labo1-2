// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package config

import (
	"sdr/labo1/src/dto"
	"sdr/labo1/src/types"
)

// UserWithPassword contains the user credentials for authentication
type UserWithPassword struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ServerConfiguration contains the information
type ServerConfiguration struct {
	Id            int                `json:"-"`
	Servers       []string           `json:"servers"`
	Users         []UserWithPassword `json:"users"`
	Events        []dto.Event        `json:"events"`
	Debug         bool               `json:"debug"`
	ShowInfosLogs bool               `json:"showInfosLogs"`
}

// FullUrl gets the formatted connection URL
func (config ServerConfiguration) FullUrl() string {
	return config.Servers[config.Id]
}

func (config ServerConfiguration) GetOtherServers() []string {
	return append(config.Servers[0:config.Id], config.Servers[config.Id+1:]...)
}

// GetData Get the users and events from a ServerConfiguration
func (config ServerConfiguration) GetData() (users map[int]*types.User, events []*types.Event) {
	users = make(map[int]*types.User)
	for _, user := range config.Users {
		users[user.Id] = &types.User{
			Id:       user.Id,
			Username: user.Username,
			Password: user.Password,
		}
	}

	for _, event := range config.Events {
		e := &types.Event{
			Id:           event.Id,
			Name:         event.Name,
			Open:         event.Open,
			OrganizerId:  event.Organizer.Id,
			Jobs:         make(map[int]*types.Job),
			Participants: make(map[int]int),
		}
		for _, job := range event.Jobs {
			e.Jobs[job.Id] = &types.Job{
				Id:       job.Id,
				Name:     job.Name,
				Capacity: job.Capacity,
			}
		}
		for _, participant := range event.Participants {
			e.Register(participant.User.Id, participant.JobId)
		}
		events = append(events, e)
	}
	return
}
