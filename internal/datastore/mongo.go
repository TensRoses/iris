package datastore

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/globalsign/mgo"

	"github.com/globalsign/mgo/bson"
)

func initMgoSessions(dbname, ip string) {
	dbLogger.Info("attempting to connect to mongo via MONGO_ADDR")

	// https://stackoverflow.com/a/42522753/8643197.
	// here we pass ip as replica sets.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	dialInfo, _ := mgo.ParseURL(ip)
	dialInfo.Timeout = 5 * time.Second
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, er := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, er
	}

	Session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		dbLogger.Error(fmt.Errorf("error while establishing connection with mongo: %w", err))
	}

	users = Session.DB(dbname).C("users")
}

func insert(user User) error {
	return users.Insert(user)
}

func fetch(discordID string) (User, error) {
	var u User

	err := users.Find(bson.M{"discordid": discordID}).One(&u)
	return u, err
}

func update(discordID string, mins int) error {
	u, _ := fetch(discordID)

	newMin := u.MinutesStudied + mins

	err := users.Update(bson.M{"discordid": u.DiscordId}, bson.M{"$set": bson.M{"minutesstudied": newMin}})
	return err
}
