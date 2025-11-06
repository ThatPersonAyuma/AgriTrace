package core

import (
	"AgriTrace/Internal/Adapters/Log"
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
	"fmt"
)

func LoginCheck(usrname, pw, id string, users map[string]string, job_store *generic.JobStore) func() error {
	return func() error{
		d, ok := job_store.Data[id]
		if !ok {
			return fmt.Errorf("Work Id tidak ada %s", id)
		}
		value, ok := users[usrname]
		var err error
		if ok{
			if pw == value {
				d.Status = "Done"
				d.Result = map[string]any{"logged_in": true, "reason": fmt.Sprintf("Selamat Datang %s", usrname)}
				err = log.CreateLog("login berhasil atas username: " + usrname, "Success")()
			}else{
				d.Status = "Done"
				d.Result = map[string]any{"logged_in": false, "reason": "Password Salah"}
				err = log.CreateLog("login gagal atas username: " + usrname + " password salah", "Wrong Password")()
			}
		}else{
			d.Status = "Done"
			d.Result = map[string]any{"logged_in": false, "reason": "Username tidak ada"}
			err = log.CreateLog("username tidak ada", "Not Found")()
		}
		job_store.Data[id] = d
		fmt.Println(id)
		return err
	}
}

func ListenLogin(b *event_bus.EventBus, topic, worker_topic string, job_store *generic.JobStore){
	sub := b.Subscribe(topic)
	go func(job_store *generic.JobStore){
		GetUsers := func() map[string]string {return map[string]string{
								"admin":"admin123",
								"bayu":"bayu123",
							}}
		for event := range sub{
			userLogin, ok := event.Payload.(generic.UserLogin)
			if ok {
				work := LoginCheck(userLogin.Username, userLogin.Password,
					event.WorkId, GetUsers(), job_store)
				b.Publish(worker_topic, event_bus.Event{WorkId: event.WorkId, Payload: work})
			}
		}
	}(job_store)
}