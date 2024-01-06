package main

import (
	"fmt"
	"os"
)

func ResurfaceByScore(g *GameState, s *State) bool {
	// fmt.Fprintf(os.Stderr, "my score: %d foe score: %d\n", s.MyScore, s.FoeScore)
	myScoreMatrix := [][]int{
		[]int{0, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 0},
	}
	for _, drone := range s.MyDrones {
		for c := range s.DroneScnas[drone.ID] {
			gs := g.GetCreature(c)
			myScoreMatrix[gs.Color][gs.Type] = gs.Type + 1
			// fmt.Fprintf(os.Stderr, "d: %d scan: %d\n", drone.ID, c)
		}
	}
	foeScoreMatrix := [][]int{
		[]int{0, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 0},
		[]int{0, 0, 0},
	}
	for _, drone := range s.FoeDrones {
		for c := range s.DroneScnas[drone.ID] {
			gs := g.GetCreature(c)
			foeScoreMatrix[gs.Color][gs.Type] = gs.Type + 1
			// fmt.Fprintf(os.Stderr, "d: %d scan: %d\n", drone.ID, c)
		}
	}
	myResurfaceScore, myAdditionResurfaceForTheFirst := matrixScore(myScoreMatrix)
	foeResurfaceScore, foeAdditionResurfaceForTheFirst := matrixScore(foeScoreMatrix)
	myTotal := s.MyScore + myResurfaceScore + myAdditionResurfaceForTheFirst
	foeTotal := s.FoeScore + foeResurfaceScore + foeAdditionResurfaceForTheFirst
	fmt.Fprintf(os.Stderr, "calc my score: %d+%d+%d=%d\n",
		s.MyScore, myResurfaceScore, myAdditionResurfaceForTheFirst, myTotal)
	fmt.Fprintf(os.Stderr, "calc foe score: %d+%d+%d=%d\n",
		s.FoeScore, foeResurfaceScore, foeAdditionResurfaceForTheFirst, foeTotal)
	return myAdditionResurfaceForTheFirst > 0 && myTotal > foeResurfaceScore && myTotal > 42 // :-)
}

func matrixScore(matrix [][]int) (resurfaceScore int, additionResurfaceForTheFirst int) {
	resurfaceScore = 0
	forTheFirst := []int{}
	for _, k := range matrix {
		if k[0] != 0 && k[1] != 0 && k[2] != 0 {
			resurfaceScore += len(k)
			forTheFirst = append(forTheFirst, len(k))
		}
	}
	for _type := 0; _type < 3; _type++ {
		for _color := 0; _color < 4; _color++ {
			resurfaceScore += matrix[_color][_type]
			forTheFirst = append(forTheFirst, matrix[_color][_type])
		}
		if matrix[0][_type] != 0 &&
			matrix[1][_type] != 0 &&
			matrix[2][_type] != 0 &&
			matrix[3][_type] != 0 {
			resurfaceScore += len(matrix)
			forTheFirst = append(forTheFirst, len(matrix))
		}
	}
	additionResurfaceForTheFirst = 0
	for _, s := range forTheFirst {
		additionResurfaceForTheFirst += s
	}
	return
}
