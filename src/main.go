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
				posX := drone.X
				if drone.X < MaxPosistionX/2 {
					posX = 2000
				} else {
					posX = 8000
				}
				// game.MoveDrone(drone, Point{X: drone.X, Y: 2500})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
				game.MoveDrone(drone, Point{X: posX, Y: 8500})
				game.MoveDrone(drone, Point{X: posX, Y: MaxPosistionY - int(AutoScanDistance)})
				game.MoveDrone(drone, Point{X: posX, Y: 0})
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
			for _, s := range s.Creatures {
				gs := game.GetCreature(s.ID)
				if gs.Type < 0 && drone.DistanceToPoint(s.Point()) < AutoScanDistance {
					drone.NearMonster = true
				}
			}

			drone.TurnLight(game)
			drone.SolveRadarRadius(game, s.MapRadar[drone.ID])
			drone.DebugRadarRadius()

			newPoint := game.FirstCommand(drone)
			if drone.DistanceToPoint(newPoint) < SurfaceDistance {
				newPoint = game.PopCommand(drone)
			}
			m := drone.Solve(game, s, s.MapRadar[drone.ID], newPoint)
			DebugLocation(m, drone.ID, newPoint)
			v := drone.SolveToGraph(game, s, m, newPoint)
			BFS(v, func(v *Vertex) {
				// fmt.Fprintf(os.Stderr, ">>(%d:%d:%d)\n", v.ID.X, v.ID.Y, len(v.Vertices))
				for k := range v.Vertices {
					v.Vertices[k].Node.Score += v.Node.Score
				}
			})
			DebugLocation(m, drone.ID, newPoint)
			// DebugVertex(v)

			drone.MoveByVertex(v)
			// _ = drone.MoveByLocation(m, nil)
			// DebugLocation(m, drone.ID, newPoint)
		}
	}
}
