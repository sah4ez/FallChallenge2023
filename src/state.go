package main

import (
	"fmt"
	"os"
	"sort"
)

type State struct {
	MyScore  int
	FoeScore int

	MyScanCount int
	MyCreatures map[int]struct{}

	FoeScanCount int
	FoeCreatures []int

	MyDroneCount int
	MyDrones     []Drone

	FoeDroneCount int
	FoeDrones     []Drone

	DroneScanCount int
	DroneScnas     map[int]map[int]struct{}

	VisibleCreatureCount int
	Creatures            []Creature

	RadarBlipCount     int
	Radar              []Radar
	MapRadar           map[int][]Radar
	DroneCreatureRadar map[int]map[int]string
	AvailableCreatures map[int]struct{}
}

func (s *State) DebugRadar() {
	fmt.Fprintf(os.Stderr, "got radar: ")
	for _, r := range s.Radar {
		fmt.Fprintf(os.Stderr, "(%s %d %d),", r.Radar, r.DroneID, r.CreatureID)
	}
	fmt.Fprintln(os.Stderr)
}

func (s *State) DebugRadarByDroneID(droneID int) {
	var d Drone
	for _, d = range s.MyDrones {
		if d.ID == droneID {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "got radar %d (%s): ", droneID, d.DetectMode())
	for _, r := range s.MapRadar[droneID] {
		fmt.Fprintf(os.Stderr, "(%s %d),", r.Radar, r.CreatureID)
	}
	fmt.Fprintln(os.Stderr)
}

func (s *State) DebugVisibleCreatures() {
	fmt.Fprintf(os.Stderr, "Visible: ")
	for _, c := range s.Creatures {
		fmt.Fprintf(os.Stderr, "(%d %d %d %d %d),", c.ID, c.X, c.Y, c.Vx, c.Vy)
	}
	fmt.Fprintln(os.Stderr)
}

func (s *State) CheckCreatureID(droneID int, creatureID int) bool {
	v, ok := s.MapRadar[droneID]
	if !ok {
		return ok
	}
	for _, r := range v {
		if r.CreatureID == creatureID {
			return true
		}
	}
	return false
}

func (s *State) CreaturesAllScannedOrSaved() bool {
	hash := map[int]struct{}{}
	for _, d := range s.MyDrones {
		if v, ok := s.DroneScnas[d.ID]; ok {
			for k := range v {
				hash[k] = struct{}{}
			}
		}
	}
	allScanned := true
	for k := range s.MyCreatures {
		if _, ok := s.AvailableCreatures[k]; !ok {
			allScanned = false
			break
		}
	}
	fmt.Fprintf(os.Stderr, "%d %d %d\n", len(s.AvailableCreatures), len(hash), len(s.MyCreatures))
	if !allScanned {
		return false
	}
	return len(s.AvailableCreatures) <= (len(hash) + len(s.MyCreatures))
}

func NewState(g *GameState) *State {
	s := &State{
		MyCreatures:        make(map[int]struct{}, 0),
		FoeCreatures:       make([]int, 0),
		MyDrones:           make([]Drone, 0),
		FoeDrones:          make([]Drone, 0),
		DroneScnas:         make(map[int]map[int]struct{}, 0),
		Creatures:          make([]Creature, 0),
		Radar:              make([]Radar, 0),
		MapRadar:           make(map[int][]Radar, 0),
		DroneCreatureRadar: make(map[int]map[int]string, 0),
		AvailableCreatures: make(map[int]struct{}, 0),
	}

	fmt.Scan(&s.MyScore)
	fmt.Scan(&s.FoeScore)

	fmt.Scan(&s.MyScanCount)
	for i := 0; i < s.MyScanCount; i++ {
		var creatureId int
		fmt.Scan(&creatureId)
		s.MyCreatures[creatureId] = struct{}{}
	}

	fmt.Scan(&s.FoeScanCount)
	for i := 0; i < s.FoeScanCount; i++ {
		var creatureId int
		fmt.Scan(&creatureId)
		s.FoeCreatures = append(s.FoeCreatures, creatureId)
	}

	fmt.Scan(&s.MyDroneCount)
	for i := 0; i < s.MyDroneCount; i++ {
		d := NewDrone()
		s.MyDrones = append(s.MyDrones, d)
		g.AddDroneCounts(d.ID)
	}

	fmt.Scan(&s.FoeDroneCount)
	for i := 0; i < s.FoeDroneCount; i++ {
		d := NewDrone()
		s.FoeDrones = append(s.FoeDrones, d)
	}

	fmt.Scan(&s.DroneScanCount)
	for i := 0; i < s.DroneScanCount; i++ {
		var droneId, creatureId int
		fmt.Scan(&droneId, &creatureId)
		if _, ok := s.DroneScnas[droneId]; ok {
			s.DroneScnas[droneId][creatureId] = struct{}{}
		} else {
			s.DroneScnas[droneId] = map[int]struct{}{creatureId: struct{}{}}
		}
	}

	fmt.Scan(&s.VisibleCreatureCount)

	for i := 0; i < s.VisibleCreatureCount; i++ {
		c := NewCreature()
		s.Creatures = append(s.Creatures, c)
	}
	sort.SliceStable(s.Creatures, func(i, j int) bool {
		ir := g.MapCreatures[s.Creatures[i].ID]
		jr := g.MapCreatures[s.Creatures[j].ID]
		return ir.Type < jr.Type
	})

	fmt.Scan(&s.RadarBlipCount)

	for i := 0; i < s.RadarBlipCount; i++ {
		r := NewRadar()
		s.Radar = append(s.Radar, r)
		if v, ok := s.MapRadar[r.DroneID]; !ok {
			s.MapRadar[r.DroneID] = []Radar{r}
		} else {
			v = append(v, r)
			s.MapRadar[r.DroneID] = v
		}
		if v, ok := s.DroneCreatureRadar[r.DroneID]; !ok {
			s.DroneCreatureRadar[r.DroneID] = map[int]string{r.CreatureID: r.Radar}
		} else {
			v[r.CreatureID] = r.Radar
			s.DroneCreatureRadar[r.DroneID] = v
		}

		if g.GetCreature(r.CreatureID).Type >= 0 {
			s.AvailableCreatures[r.CreatureID] = struct{}{}
		}
	}
	sort.SliceStable(s.Radar, func(i, j int) bool {
		ir := g.MapCreatures[s.Radar[i].CreatureID]
		jr := g.MapCreatures[s.Radar[j].CreatureID]
		return ir.Type < jr.Type
	})

	return s
}
