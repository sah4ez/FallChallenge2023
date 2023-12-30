package main

func main() {
	game := NewGame()

	game.LoadCreatures()

	for {
		s := game.LoadState()

		for _, d := range s.MyDrones {
			d.TurnLight()
			d.Wait()
		}
	}
}
