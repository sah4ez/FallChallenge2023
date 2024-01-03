package main

import (
	"fmt"
	"os"
	"sync"
)

func main() {
	game := NewGame()

	game.LoadCreatures()
	game.DebugCreatures()

	var once sync.Once
	var onceDrone sync.Once

	for {
		s := game.LoadState()
		// s.DebugRadar()
		s.DebugVisibleCreatures()

		once.Do(game.DebugLightCoutns)
		onceDrone.Do(func() {
			for i := range s.MyDrones {
				drone := s.MyDrones[i]
				game.MoveDrone(drone, Point{X: drone.X, Y: 2500})
				game.MoveDrone(drone, Point{X: drone.X, Y: 8500})
				game.MoveDrone(drone, Point{X: int(MaxPosistionX / 2), Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: int(MaxPosistionX / 2), Y: SurfaceDistance})
			}
		})

		for i := range s.MyDrones {
			drone := s.MyDrones[i]
			if drone.IsEmergency() {
				game.RemoveDroneTarget(drone)
				fmt.Fprintln(os.Stderr, drone.ID, "emergency", s.DroneScnas[drone.ID])
				drone.Wait()
				continue
			}

			drone.TurnLight(game)
			drone.SolveRadarRadius(game, s.MapRadar[drone.ID])
			drone.DebugRadarRadius()

			newPoint := game.FirstCommand(drone)
			m := drone.Solve(game, s, s.MapRadar[drone.ID], newPoint)
			DebugLocation(m, drone.ID, newPoint)
			if drone.DistanceToPoint(newPoint) < AutoScanDistance {
				drone.TurnLight(game)
				newPoint = game.PopCommand(drone)
			}
			drone.TurnLight(game)
			drone.Move(newPoint)
		}
	}
}
