package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"
)

type Zone struct {
	ID       int
	Owner    int
	PODS     [4]int
	Platinum int

	Continent int
	Neighbors []*Zone
}

func (z Zone) String() string {
	return "Zone[" + strconv.Itoa(z.ID) + "](O:" + strconv.Itoa(z.Owner) + " V:" + strconv.Itoa(z.Platinum) + " C:" + strconv.Itoa(z.Continent) + " PODS:" + strconv.Itoa(z.PODS[0]) + "," + strconv.Itoa(z.PODS[1]) + "," + strconv.Itoa(z.PODS[2]) + "," + strconv.Itoa(z.PODS[3]) + ")"
}

func (z Zone) IsSpawnable(player int) bool {
	return z.Owner == -1 || z.Owner == player
}

type RandomZone map[int]*Zone

func (r RandomZone) PlayerPOD(player int) *Zone {
	owned := []*Zone{}
	for _, node := range r {
		if node.PODS[player] > 0 {
			owned = append(owned, node)
		}
	}
	return owned[int(rand.Int31n(int32(len(owned))))]
}

func (r RandomZone) EnemyPOD(player int) *Zone {
	owned := []*Zone{}
	for _, node := range r {
		for i := 0; i < 4; i++ {
			if i != player && node.PODS[i] > 0 {
				owned = append(owned, node)
			}
		}
	}
	return owned[int(rand.Int31n(int32(len(owned))))]
}

func (r RandomZone) Spawnable(player int) *Zone {
	spawnable := []*Zone{}
	for _, node := range r {
		if node.Owner == -1 || node.Owner == player {
			spawnable = append(spawnable, node)
		}
	}
	if len(spawnable) != 0 {
		return spawnable[int(rand.Int31n(int32(len(spawnable))))]
	}
	return nil
}

type Continent struct {
	ID    int
	Zones map[int]*Zone
}

func (c Continent) Size() int {
	return len(c.Zones)
}

func (c *Continent) FriendlyCount(player int) int {
	sum := 0
	for _, node := range c.Zones {
		sum += node.PODS[player]
	}
	return sum
}

func (c *Continent) EnemyCount(player int) int {
	sum := 0
	for _, node := range c.Zones {
		for i := 0; i < 4; i++ {
			if i != player {
				sum += node.PODS[i]
			}
		}
	}
	return sum
}

type World struct {
	Zones      map[int]*Zone
	Continents map[int]*Continent

	RoundNumber int
	PlayerID    int
	Platinum    int

	MoveMessage  string
	SpawnMessage string
}

func (w *World) AddMove(number, start, end int) {
	if w.MoveMessage == "WAIT" {
		w.MoveMessage = ""
	}
	w.MoveMessage += strconv.Itoa(number) + " " + strconv.Itoa(start) + " " + strconv.Itoa(end) + " "
}

func (w World) AvailableSpawns() int {
	return w.Platinum / 20
}

func (w *World) AddSpawn(number, position int) {
	if w.SpawnMessage == "WAIT" {
		w.SpawnMessage = ""
	}

	w.Platinum -= 20
	w.SpawnMessage += strconv.Itoa(number) + " " + strconv.Itoa(position) + " "
}

func (w *World) Step() {
	w.RoundNumber++

	fmt.Println(w.MoveMessage)
	fmt.Println(w.SpawnMessage)

	w.MoveMessage = "WAIT"
	w.SpawnMessage = "WAIT"
}

func (w *World) CalculateContinents() {
	continent := 0
	visited := make([]bool, len(w.Zones))
	for i := 0; i < len(visited); i++ {
		if !visited[i] {
			w.SetContinentBFS(continent, w.Zones[i], visited)
			continent++
		}
	}

	// Setup Continents
	w.Continents = make(map[int]*Continent)
	for i := 0; i < continent; i++ {
		w.Continents[i] = &Continent{ID: i, Zones: make(map[int]*Zone)}
	}

	// Fill Continents
	for id, zone := range w.Zones {
		w.Continents[zone.Continent].Zones[id] = zone
	}
}

func (w *World) SetContinentBFS(continent int, zone *Zone, visited []bool) {
	if visited[zone.ID] {
		return
	}
	visited[zone.ID] = true

	zone.Continent = continent
	for _, neighbor := range zone.Neighbors {
		w.SetContinentBFS(continent, neighbor, visited)
	}
}

func (w World) Path(start Zone, endTest func(*Zone, int) bool) []int {
	distance := make(map[int]int)
	previous := make(map[int]int)
	nodes := make(map[int]*Zone)

	path := func(target *Zone, previous map[int]int) []int {
		path := []int{}
		if target != nil {
			u := target.ID
			for {
				if previous[u] == -1 {
					break
				}

				path = append([]int{u}, path...)
				u = previous[u]
			}
		}
		return path
	}

	for _, node := range w.Continents[start.Continent].Zones {
		distance[node.ID] = math.MaxInt32
		previous[node.ID] = -1
		nodes[node.ID] = node
	}
	distance[start.ID] = 0

	for len(nodes) > 0 {
		smallest_id, smallest_dist := -1, 9999
		for _, node := range nodes {
			if distance[node.ID] < smallest_dist {
				smallest_id = node.ID
				smallest_dist = distance[node.ID]
			}
		}

		currentNode := nodes[smallest_id]
		delete(nodes, currentNode.ID)

		if endTest(currentNode, distance[currentNode.ID]) {
			return path(currentNode, previous)
		}

		for _, neighbor := range currentNode.Neighbors {
			alt := distance[currentNode.ID]

			// Favor enemy zones over unclaimed over owned
			if neighbor.Owner != start.Owner {
				if neighbor.Owner != -1 {
					alt += 1
				} else {
					alt += 2
				}
			} else {
				alt += 3
			}

			if alt < distance[neighbor.ID] {
				distance[neighbor.ID] = alt
				previous[neighbor.ID] = currentNode.ID
			}
		}
	}

	return []int{}
}

func (w *World) SpawnRandom() {
	for w.AvailableSpawns() > 0 {
		w.AddSpawn(1, RandomZone(w.Zones).Spawnable(w.PlayerID).ID)
	}
}

func (w *World) SpawnOneContinent(continent int) {
	for w.AvailableSpawns() > 0 {
		w.AddSpawn(1, RandomZone(w.Continents[continent].Zones).Spawnable(w.PlayerID).ID)
	}
}

func (w *World) SpawnRandomUnclaimedFirst() {
	empty, owned := make(map[int]*Zone), make(map[int]*Zone)
	for _, zone := range w.Zones {
		if zone.Owner == -1 {
			empty[zone.ID] = zone
		}
		if zone.Owner != -1 && zone.Owner == w.PlayerID {
			owned[zone.ID] = zone
		}
	}

	for w.AvailableSpawns() > 0 {
		zone := RandomZone(empty).Spawnable(w.PlayerID)
		if zone == nil {
			break
		}
		w.AddSpawn(1, zone.ID)
	}

	for w.AvailableSpawns() > 0 {
		w.AddSpawn(1, RandomZone(owned).Spawnable(w.PlayerID).ID)
	}
}

type BySize []*Continent

func (b BySize) Len() int           { return len(b) }
func (b BySize) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b BySize) Less(i, j int) bool { return b[i].Size() < b[j].Size() }

func (w *World) SpawnBalancePODS() {
	continents := []*Continent{}
	for _, continent := range w.Continents {
		continents = append(continents, continent)
	}
	sort.Sort(BySize(continents))

	for _, c := range continents {
		diff := (w.Continents[c.ID].EnemyCount(w.PlayerID) - w.Continents[c.ID].FriendlyCount(w.PlayerID)) + 1
		if diff > 0 {
			for i := 0; i < diff; i++ {
				if w.AvailableSpawns() == 0 {
					break
				}
				zone := RandomZone(w.Continents[c.ID].Zones).Spawnable(w.PlayerID)
				if zone == nil {
					break
				}
				w.AddSpawn(1, zone.ID)
			}
		}
	}
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Lshortfile)

	world := World{MoveMessage: "WAIT", SpawnMessage: "WAIT", Zones: make(map[int]*Zone), Continents: make(map[int]*Continent)}

	var playerCount, zoneCount, linkCount int
	fmt.Scan(&playerCount, &world.PlayerID, &zoneCount, &linkCount)

	rand.Seed(time.Now().Unix() * int64(world.PlayerID))

	// Setup World
	for i := 0; i < zoneCount; i++ {
		var id, value int
		fmt.Scan(&id, &value)

		world.Zones[id] = &Zone{ID: id, Continent: -1, Platinum: value}
	}

	for i := 0; i < linkCount; i++ {
		var zone1, zone2 int
		fmt.Scan(&zone1, &zone2)

		world.Zones[zone1].Neighbors = append(world.Zones[zone1].Neighbors, world.Zones[zone2])
		world.Zones[zone2].Neighbors = append(world.Zones[zone2].Neighbors, world.Zones[zone1])
	}
	world.CalculateContinents()

	// Handle Steps
	for {
		start := time.Now()

		fmt.Scan(&world.Platinum)

		var myUnits []*Zone
		for i := 0; i < zoneCount; i++ {
			var id, owner, podsP0, podsP1, podsP2, podsP3 int
			fmt.Scan(&id, &owner, &podsP0, &podsP1, &podsP2, &podsP3)

			world.Zones[id].Owner = owner
			world.Zones[id].PODS = [4]int{podsP0, podsP1, podsP2, podsP3}

			if world.Zones[id].PODS[world.PlayerID] > 0 {
				myUnits = append(myUnits, world.Zones[id])
			}
		}

		// Movement
		target := make(map[int]bool)
		for _, zone := range myUnits {
			units := zone.PODS[world.PlayerID]
			for units > 0 {
				path := world.Path(*zone, func(z *Zone, d int) bool {
					if !target[zone.ID] && z.Owner != world.PlayerID {
						return true
					}
					return false
				})

				if len(path) == 0 {
					break
				}
				target[path[0]] = true

				world.AddMove(1, zone.ID, path[0])
				units -= 1
			}

			for units > 0 {
				path := world.Path(*zone, func(z *Zone, _ int) bool {
					if !target[zone.ID] && z.Platinum > 0 && z.Owner != world.PlayerID {
						return true
					}
					return false
				})

				if len(path) == 0 {
					break
				}
				target[path[0]] = true

				world.AddMove(1, zone.ID, path[0])
				units -= 1
			}
		}

		world.SpawnRandom()
		world.Step()

		log.Println("Time:", time.Now().Sub(start))
	}
}
