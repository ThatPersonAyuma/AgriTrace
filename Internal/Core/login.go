package core

import (
	"AgriTrace/Internal/Adapters"
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
)

func LoginCheck(usrname string, pw string, users map[string]string) generic.Result[bool, error]{
	value, ok := users[usrname]
	if ok{
		if pw == value {
			return generic.Result[bool, error]{Value: true, Effect: adapters.CreateLog("login berhasil atas username: " + usrname, "Success")}
		}else{
			return generic.Result[bool, error]{Value: false, Effect: adapters.CreateLog("login gagal atas username: " + usrname + " password salah", "Wrong Password")}
		}
	}else{
		return generic.Result[bool, error]{Value: false, Effect: adapters.CreateLog("username tidak ada", "Not Found")}
	}
}

func ListenLogin(b *event_bus.EventBus){
	sub := b.Subscribe("Login")
	go func(){
		GetUsers := func() map[string]string {return map[string]string{
								"admin":"admin123",
								"bayu":"bayu123",
							}}
		for event := range sub{
			userLogin, ok := event.Payload.(generic.UserLogin)
			
			if ok {
				result := LoginCheck(userLogin.Username, userLogin.Password,GetUsers())
				b.Publish("Works", event_bus.Event{Payload: result.Effect})
			}
		}
	}()
}