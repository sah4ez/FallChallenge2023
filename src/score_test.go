package main

import "testing"

func TestResurfaceByScoreOnlyScan(t *testing.T) {
	gameCreatures := []*GameCreature{
		&GameCreature{ID: 4, Type: 0, Color: 0},
		&GameCreature{ID: 5, Type: 0, Color: 1},
		&GameCreature{ID: 6, Type: 0, Color: 2},
		&GameCreature{ID: 7, Type: 0, Color: 3},
		&GameCreature{ID: 8, Type: 1, Color: 0},
		&GameCreature{ID: 9, Type: 1, Color: 1},
		&GameCreature{ID: 10, Type: 1, Color: 2},
		&GameCreature{ID: 11, Type: 1, Color: 3},
		&GameCreature{ID: 12, Type: 2, Color: 0},
		&GameCreature{ID: 13, Type: 2, Color: 1},
		&GameCreature{ID: 14, Type: 2, Color: 2},
		&GameCreature{ID: 15, Type: 2, Color: 3},
	}
	mapCreatures := map[int]*GameCreature{}
	for _, c := range gameCreatures {
		mapCreatures[c.ID] = c
	}
	g := &GameState{
		Creatures:    gameCreatures,
		MapCreatures: mapCreatures,
	}
	s := &State{
		MyScore:  0,
		FoeScore: 0,
		MyDrones: []Drone{
			{ID: 0},
			{ID: 2},
		},
		DroneScnas: map[int]map[int]struct{}{
			0: map[int]struct{}{
				4:  {},
				5:  {},
				6:  {},
				7:  {},
				8:  {},
				9:  {},
				10: {},
				11: {},
				12: {},
				13: {},
				14: {},
				15: {},
			},
			1: map[int]struct{}{
				4:  {},
				5:  {},
				6:  {},
				7:  {},
				8:  {},
				9:  {},
				10: {},
				11: {},
				12: {},
				13: {},
				14: {},
				15: {},
			},
			2: map[int]struct{}{
				4:  {},
				5:  {},
				6:  {},
				7:  {},
				8:  {},
				9:  {},
				10: {},
				11: {},
				12: {},
				13: {},
				14: {},
				15: {},
			},
			3: map[int]struct{}{
				4:  {},
				5:  {},
				6:  {},
				7:  {},
				8:  {},
				9:  {},
				10: {},
				11: {},
				12: {},
				13: {},
				14: {},
				15: {},
			},
		},
	}
	ResurfaceByScore(g, s)
}
