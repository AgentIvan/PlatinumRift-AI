package main

import "fmt"
import "math/rand"
import "os"
import "strconv"
import "time"
import "log"

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

func (z Zone) UnclaimedNeighbors() []*Zone {
	var retVal []*Zone

	for _, neighbor := range z.Neighbors {
		if neighbor.Owner == -1 {
			retVal = append(retVal, neighbor)
		}
	}

	return retVal
}

func (z Zone) OwnedNeighbors(owner int) []*Zone {
	var retVal []*Zone
	for _, neighbor := range z.Neighbors {
		if neighbor.Owner == owner {
			retVal = append(retVal, neighbor)
		}
	}
	return retVal
}

func (z Zone) DefeatableNeighbors(player int) []*Zone {
	var retVal []*Zone

	for _, neighbor := range z.Neighbors {
		if neighbor.Owner != -1 && neighbor.Owner != player {
			if z.PODS[player] > neighbor.PODS[neighbor.Owner] {
				retVal = append(retVal, neighbor)
			}
		}
	}

	return retVal
}

func (z Zone) IsSpawnable(player int) bool {
	return z.Owner == -1 || z.Owner == player
}

type RandomZone []*Zone

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
	return spawnable[int(rand.Int31n(int32(len(spawnable))))]
}

type Continent struct {
	ID            int
	Zones         []*Zone
	PlatinumZones []*Zone
}

func (c Continent) Size() int {
	return len(c.Zones)
}

type World struct {
	Zones      []*Zone
	Continents []*Continent

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
	for i := 0; i < continent; i++ {
		w.Continents = append(w.Continents, &Continent{ID: i})
	}

	// Fill Continents
	for i := 0; i < len(w.Zones); i++ {
		w.Continents[w.Zones[i].Continent].Zones = append(w.Continents[w.Zones[i].Continent].Zones, w.Zones[i])
		if w.Zones[i].Platinum > 0 {
			w.Continents[w.Zones[i].Continent].PlatinumZones = append(w.Continents[w.Zones[i].Continent].PlatinumZones, w.Zones[i])
		}
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

func (w World) Path(start Zone, endTest func(*Zone) bool) []int {
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
		distance[node.ID] = len(w.Continents[start.Continent].Zones)
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

		if endTest(currentNode) {
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
	for i := 0; i < w.AvailableSpawns(); i++ {
		w.AddSpawn(1, RandomZone(w.Zones).Spawnable(w.PlayerID).ID)
	}
}

func (w *World) SpawnOneContinent(continent int) {
	for i := 0; i < w.AvailableSpawns(); i++ {
		w.AddSpawn(1, RandomZone(w.Continents[continent].Zones).Spawnable(w.PlayerID).ID)
	}
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime | log.Lshortfile)

	world := World{MoveMessage: "WAIT", SpawnMessage: "WAIT"}

	var playerCount, zoneCount, linkCount int
	fmt.Scan(&playerCount, &world.PlayerID, &zoneCount, &linkCount)

	rand.Seed(time.Now().Unix() * int64(world.PlayerID))

	// Setup World
	for i := 0; i < zoneCount; i++ {
		var id, value int
		fmt.Scan(&id, &value)

		world.Zones = append(world.Zones, &Zone{ID: id, Continent: -1, Platinum: value})
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
		for _, zone := range myUnits {
			path := world.Path(*zone, func(z *Zone) bool {
				if z.Platinum > 0 && z.Owner != world.PlayerID {
					log.Println("PTarget:", z, "\n")
					return true
				}
				return false
			})

			if len(path) > 0 {
				world.AddMove(zone.PODS[world.PlayerID], zone.ID, path[0])
				zone.PODS[world.PlayerID] -= zone.PODS[world.PlayerID]
			}

			if zone.PODS[world.PlayerID] > 0 {
				path := world.Path(*zone, func(z *Zone) bool {
					if z.Owner != world.PlayerID {
						log.Println("OTarget:", z, "\n")
						return true
					}
					return false
				})
				if len(path) > 0 {
					world.AddMove(zone.PODS[world.PlayerID], zone.ID, path[0])
					zone.PODS[world.PlayerID] -= zone.PODS[world.PlayerID]
				}
			}
		}

		world.SpawnRandom()
		world.Step()
	}
}
