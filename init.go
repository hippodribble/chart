package chart

var RoseDirections []string
var RoseLookup map[string]int

func init() {
	RoseDirections = append(RoseDirections, "N")
	RoseDirections = append(RoseDirections, "NE")
	RoseDirections = append(RoseDirections, "E")
	RoseDirections = append(RoseDirections, "SE")
	RoseDirections = append(RoseDirections, "S")
	RoseDirections = append(RoseDirections, "SW")
	RoseDirections = append(RoseDirections, "W")
	RoseDirections = append(RoseDirections, "NW")

	RoseLookup = make(map[string]int)
	RoseLookup["N"] = 0
	RoseLookup["NE"] = 1
	RoseLookup["E"] = 2
	RoseLookup["SE"] = 3
	RoseLookup["S"] = 4
	RoseLookup["SW"] = 5
	RoseLookup["W"] = 6
	RoseLookup["NW"] = 7
}
