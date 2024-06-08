package client

func RoomExist(name string) bool {
	_, ok := Rooms.Load(name)
	return ok
}
