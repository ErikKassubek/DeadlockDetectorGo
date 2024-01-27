package analysis

import (
	"analyzer/logging"
	"sort"
)

func checkForDoneBeforeAdd(routine int, id int, delta int, pos string, vc VectorClock) {
	if delta > 0 {
		checkForDoneBeforeAddAdd(routine, id, pos, vc, delta)
	} else if delta < 0 {
		checkForDoneBeforeAddDone(routine, id, pos, vc)
	} else {
		// checkForImpossibleWait(routine, id, pos, vc)
	}
}

func checkForDoneBeforeAddAdd(routine int, id int, pos string, vc VectorClock, delta int) {
	// if necessary, create maps and lists
	if _, ok := addVcs[id]; !ok {
		addVcs[id] = make(map[int][]VectorClock)
		addPos[id] = make(map[int][]string)
	}
	if _, ok := addVcs[id][routine]; !ok {
		addVcs[id][routine] = make([]VectorClock, 0)
		addPos[id][routine] = make([]string, 0)
	}

	// add the vector clock and position to the list
	for i := 0; i < delta; i++ {
		addVcs[id][routine] = append(addVcs[id][routine], vc.Copy())
		addPos[id][routine] = append(addPos[id][routine], pos)
	}
}

func checkForDoneBeforeAddDone(routine int, id int, pos string, vc VectorClock) {
	// if necessary, create maps and lists
	if _, ok := doneVcs[id]; !ok {
		doneVcs[id] = make(map[int][]VectorClock)
		donePos[id] = make(map[int][]string)
	}
	if _, ok := doneVcs[id][routine]; !ok {
		doneVcs[id][routine] = make([]VectorClock, 0)
		donePos[id][routine] = make([]string, 0)
	}

	// add the vector clock and position to the list
	doneVcs[id][routine] = append(doneVcs[id][routine], vc.Copy())
	donePos[id][routine] = append(donePos[id][routine], pos)

	// for now, test new vector clock against all add vector clocks
	// TODO: make this more efficient
	// for r, vcs := range addVcs[id] {
	// 	for i, vcAdd := range vcs {
	// 		happensBefore := GetHappensBefore(vcAdd, vc)
	// 		if happensBefore == Concurrent {
	// 			found := "Found concurrent Add and Done on same waitgroup:\n"
	// 			found += "\tdone: " + donePos[id][routine][len(donePos[id][routine])-1] + "\n"
	// 			found += "\tadd: " + addPos[id][r][i]
	// 			logging.Result(found, logging.CRITICAL)
	// 		}
	// 	}
	// }
}

/*
 * Check if a wait group counter could become negative
 * For each done operation, count the number of add operations a that happen before
 * the done operation, the number of done operations d that happen before the done operation
 * and the number of done operations d' that happen concurrent to the done operation.
 * If a < d + d', then the counter could become negative.
 * In this case, print a warning.
 */
func CheckForDoneBeforeAdd() {
	for id, vcs := range doneVcs { // for all waitgroups id
		for routine, vcs := range vcs { // for all routines
			for op, vcDone := range vcs { // for all done operations
				// count the number of add operations a that happen before or concurrent to the done operation
				countAdd := 0
				addPosList := []string{}
				for routineAdd, vcs := range addVcs[id] { // for all routines
					for opAdd, vcAdd := range vcs { // for all add operations
						happensBefore := GetHappensBefore(vcAdd, vcDone)
						if happensBefore == Before {
							countAdd++
						} else if happensBefore == Concurrent {
							addPosList = append(addPosList, addPos[id][routineAdd][opAdd])
						}
					}
				}
				// count the number of done operations d that happen before the done operation
				countDone := 0
				donePosListConc := []string{}
				donePosListBefore := []string{}
				for routine2, vcs := range doneVcs[id] { // for all routines
					for op2, vcDone2 := range vcs { // for all done operations
						if routine2 == routine && op2 == op {
							continue
						}
						happensBefore := GetHappensBefore(vcDone2, vcDone)
						if happensBefore == Before {
							countDone++
							donePosListBefore = append(donePosListBefore, donePos[id][routine2][op2])
						} else if happensBefore == Concurrent {
							countDone++
							donePosListConc = append(donePosListConc, donePos[id][routine2][op2])
						}
					}
				}

				if countAdd < countDone {
					createDoneBeforeAddMessage(id, routine, op, addPosList, donePosListConc, donePosListBefore)
				}
			}
		}
	}
}

func createDoneBeforeAddMessage(id int, routine int, op int, addPosList []string, donePosListConc []string, donePosListBefore []string) {
	uniquePos := make(map[string]bool)
	sort.Strings(addPosList)
	sort.Strings(donePosListConc)
	sort.Strings(donePosListBefore)

	found := "Possible negative waitgroup counter:\n"
	found += "\tdone: " + donePos[id][routine][op] + "\n"
	found += "\tdone/add: "

	for i, pos := range donePosListConc {
		if uniquePos[pos] {
			continue
		}
		if i != 0 {
			found += ";"
		}
		found += pos
		uniquePos[pos] = true
	}

	if len(donePosListConc) > 0 && (len(donePosListBefore) > 0 || len(addPosList) > 0) {
		found += ";"
	}

	for i, pos := range donePosListBefore {
		if uniquePos[pos] {
			continue
		}
		if i != 0 {
			found += ";"
		}
		found += pos
		uniquePos[pos] = true
	}

	if len(donePosListBefore) > 0 && len(addPosList) > 0 {
		found += ";"
	}

	for i, pos := range addPosList {
		if uniquePos[pos] {
			continue
		}
		if i != 0 {
			found += ";"
		}
		found += pos
		uniquePos[pos] = true
	}
	logging.Result(found, logging.CRITICAL)
}