package sqlite

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActors(t *testing.T) {
	var log *slog.Logger
	log=slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	db, err:= sql.Open("sqlite", "vk-films-testovoe\\filmoteka\\cmd\\main\\storage.db")
	if err != nil {
		require.NoError(t, err)
	}

	s:=&Storage{db: db}

	actor1:=Actor{
		Name: "Vova",
		Gender: "male",
		BirthDate: "16.05.2000",
		Films: []string{"Harry Potter"},
	}
	actor2:=Actor{
		Name: "Lisa",
		Gender: "female",
		BirthDate: "13.03.2001",
		Films: []string{"Harry Potter", "Fast and furious"},
	}

	err=PostActorToStorage(s,actor1)
	if err != nil {
		require.NoError(t, err)
	}
	err=PostActorToStorage(s,actor2)
	if err != nil {
		require.NoError(t, err)
	}

	actorsList, err:=GetAllActorsFromStorage(s,log)
	if err != nil {
		require.NoError(t, err)
	}

	count:=0

	for _,actors:=range(actorsList){
		if actors.Name==actor1.Name{
			actor1.ActorId=actors.ActorId
			count++
		}else if actors.Name==actor2.Name{
			actor2.ActorId=actors.ActorId
			count++
		}
	}

	if count!=2{
		require.Equal(t, count,2)
	}

	actorFromStorage, err:=GetOneActorFromStorage(s,actor1.ActorId,log)
	if err != nil {
		require.NoError(t, err)
	}

	assert.Equal(t,actor1,actorFromStorage)

	actor3:=Actor{
		Name: "Lisa",
		Gender: "female",
		BirthDate: "14.03.2001",
		Films: []string{"Harry Potter", "Fast and furious"},
	}

	err=UpdateActor(s,actor3)
	if err != nil {
		require.NoError(t, err)
	}

	actorFromStorage2, err:=GetOneActorFromStorage(s,actor2.ActorId,log)
	if err != nil {
		require.NoError(t, err)
	}

	assert.Equal(t,actor3,actorFromStorage2)

	err=DeleteActor(s,actorFromStorage2.ActorId)
	if err != nil {
		require.NoError(t, err)
	}

	_, err=GetOneActorFromStorage(s,actor2.ActorId,log)
	if err != nil {
		require.Error(t, err)
	}

	err=DeleteActor(s,actorFromStorage.ActorId)
	if err != nil {
		require.NoError(t, err)
	}

	_, err=GetOneActorFromStorage(s,actor1.ActorId,log)
	if err != nil {
		require.Error(t, err)
	}
}