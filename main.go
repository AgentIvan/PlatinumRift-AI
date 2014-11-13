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
	Platinum int

	PODS     [4]int
	UsedPODS int

	Continent int
	Neighbors []*Zone

	// Pathing Variables
	Distance map[int]int
	Previous map[int]int
}

func (z Zone) String() string {
	return "Zone[" + strconv.Itoa(z.ID) + "](O:" + strconv.Itoa(z.Owner) + " V:" + strconv.Itoa(z.Platinum) + " C:" + strconv.Itoa(z.Continent) + " PODS:" + strconv.Itoa(z.PODS[0]) + "," + strconv.Itoa(z.PODS[1]) + "," + strconv.Itoa(z.PODS[2]) + "," + strconv.Itoa(z.PODS[3]) + ")"
}

func (z Zone) IsSpawnable(player int) bool {
	return z.Owner == -1 || z.Owner == player
}

func (z *Zone) PathTo(target *Zone) []int {
	path := []int{}
	if target != nil {
		u := target.ID
		for {
			if z.Previous[u] == -1 {
				break
			}

			path = append([]int{u}, path...)
			u = z.Previous[u]
		}
	}
	return path
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

	PlatinumZones  []*Zone
	FriendlyZones  []*Zone
	EnemyZones     []*Zone
	UnclaimedZones []*Zone

	PlayerUnits []*Zone
	EnemyUnits  []*Zone

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
	w.Zones[start].UsedPODS += number
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

func (w *World) Initialize() {
	var playerCount, zoneCount, linkCount int
	fmt.Scan(&playerCount, &w.PlayerID, &zoneCount, &linkCount)

	w.MoveMessage = "WAIT"
	w.SpawnMessage = "WAIT"

	w.Zones = make(map[int]*Zone)
	w.Continents = make(map[int]*Continent)

	rand.Seed(time.Now().Unix() * int64(w.PlayerID))

	for i := 0; i < zoneCount; i++ {
		zone := Zone{Continent: -1}
		fmt.Scan(&zone.ID, &zone.Platinum)

		w.Zones[zone.ID] = &zone
		if zone.Platinum > 0 {
			w.PlatinumZones = append(w.PlatinumZones, &zone)
		}
	}

	for i := 0; i < linkCount; i++ {
		var zone1, zone2 int
		fmt.Scan(&zone1, &zone2)

		w.Zones[zone1].Neighbors = append(w.Zones[zone1].Neighbors, w.Zones[zone2])
		w.Zones[zone2].Neighbors = append(w.Zones[zone2].Neighbors, w.Zones[zone1])
	}

	w.CalculateContinents()
}

func (w *World) Update() {
	// Update Input
	fmt.Scan(&w.Platinum)
	for i := 0; i < len(w.Zones); i++ {
		var id int
		fmt.Scan(&id)
		fmt.Scan(&w.Zones[id].Owner, &w.Zones[id].PODS[0], &w.Zones[id].PODS[1], &w.Zones[id].PODS[2], &w.Zones[id].PODS[3])
	}

	// Clear Transitory Data
	w.FriendlyZones = []*Zone{}
	w.EnemyZones = []*Zone{}
	w.UnclaimedZones = []*Zone{}

	w.PlayerUnits = []*Zone{}
	w.EnemyUnits = []*Zone{}

	// Update Data
	for _, zone := range w.Zones {
		zone.UsedPODS = 0

		for i := 0; i < 4; i++ {
			if zone.PODS[i] > 0 {
				if i == w.PlayerID {
					w.PlayerUnits = append(w.PlayerUnits, zone)
					w.UpdatePathing(zone)
				} else {
					w.EnemyUnits = append(w.EnemyUnits, zone)
				}
			}
		}

		switch zone.Owner {
		case -1:
			w.UnclaimedZones = append(w.UnclaimedZones, zone)
		case w.PlayerID:
			w.FriendlyZones = append(w.FriendlyZones, zone)
		default:
			w.EnemyZones = append(w.EnemyZones, zone)
		}
	}
}

func (w *World) UpdatePathing(z *Zone) {
	if z.Distance == nil {
		z.Distance = make(map[int]int)
	}
	if z.Previous == nil {
		z.Previous = make(map[int]int)
	}

	nodes := make(map[int]*Zone)
	for _, node := range w.Continents[z.Continent].Zones {
		z.Distance[node.ID] = math.MaxInt32
		z.Previous[node.ID] = -1
		nodes[node.ID] = node
	}
	z.Distance[z.ID] = 0

	for len(nodes) > 0 {
		smallest_id, smallest_dist := -1, math.MaxInt32
		for _, node := range nodes {
			if z.Distance[node.ID] < smallest_dist {
				smallest_id = node.ID
				smallest_dist = z.Distance[node.ID]
			}
		}

		currentNode := nodes[smallest_id]
		delete(nodes, currentNode.ID)

		for _, neighbor := range currentNode.Neighbors {
			alt := z.Distance[currentNode.ID]

			// Favor enemy zones over unclaimed over owned
			if neighbor.Owner != z.Owner {
				if neighbor.Owner != -1 {
					alt += 1
				} else {
					alt += 2
				}
			} else {
				alt += 3
			}

			if alt < z.Distance[neighbor.ID] {
				z.Distance[neighbor.ID] = alt
				z.Previous[neighbor.ID] = currentNode.ID
			}
		}
	}
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

	var world World
	world.Initialize()

	// Handle Steps
	for {
		start := time.Now()
		world.Update()

		// Calculate Movement
		for _, pZone := range world.PlatinumZones {
			shortest, index := math.MaxInt32, -1
			for _, zone := range world.PlayerUnits {
				if zone.UsedPODS != zone.PODS[world.PlayerID] && zone.Continent == pZone.Continent {
					if dist, ok := zone.Distance[pZone.ID]; ok {
						if dist < shortest {
							shortest = dist
							index = zone.ID
						}
					}
				}
			}

			if index != -1 {
				if path := world.Zones[index].PathTo(pZone); len(path) > 0 {
					world.AddMove(world.Zones[index].PODS[world.PlayerID], world.Zones[index].ID, path[0])
				}
			}
		}

		// Calculate Spawns
		world.SpawnRandom()

		// Initiate Step
		world.Step()

		log.Println("Time:", time.Now().Sub(start))
	}
}
