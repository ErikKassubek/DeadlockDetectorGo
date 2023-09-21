// DEDEGO-FILE_START

package runtime

var DedegoRoutines map[uint64]*[]dedegoTraceElement
var DedegoRoutinesLock *mutex

var projectPath string

type DedegoRoutine struct {
	id    uint64
	G     *g
	Trace []dedegoTraceElement
	lock  *mutex
}

/*
 * Create a new dedego routine
 * Params:
 * 	g: the g struct of the routine
 * Return:
 * 	the new dedego routine
 */
func newDedegoRoutine(g *g) *DedegoRoutine {
	routine := &DedegoRoutine{id: GetDedegoRoutineId(), G: g,
		Trace: make([]dedegoTraceElement, 0),
		lock:  &mutex{}}

	if DedegoRoutinesLock == nil {
		DedegoRoutinesLock = &mutex{}
	}

	lock(DedegoRoutinesLock)
	defer unlock(DedegoRoutinesLock)

	if DedegoRoutines == nil {
		DedegoRoutines = make(map[uint64]*[]dedegoTraceElement)
	}

	DedegoRoutines[routine.id] = &routine.Trace // Todo: causes warning in race detector

	return routine
}

/*
 * Add an element to the trace of the current routine
 * Params:
 * 	elem: the element to add
 * Return:
 * 	the index of the element in the trace
 */
func (gi *DedegoRoutine) addToTrace(elem dedegoTraceElement) int {
	// do nothing if tracer disabled
	if dedegoDisabled {
		return -1
	}
	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}
	lock(gi.lock)
	defer unlock(gi.lock)
	if gi.Trace == nil {
		gi.Trace = make([]dedegoTraceElement, 0)
	}
	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *DedegoRoutine) getElement(index int) dedegoTraceElement {
	lock(gi.lock)
	defer unlock(gi.lock)
	return gi.Trace[index]
}

/*
 * Update an element in the trace of the current routine
 * Params:
 * 	index: the index of the element to update
 * 	elem: the new element
 */
func (gi *DedegoRoutine) updateElement(index int, elem dedegoTraceElement) {
	if dedegoDisabled {
		return
	}

	if gi == nil {
		return
	}

	if gi.Trace == nil {
		panic("Tried to update element in nil trace")
	}

	if index >= len(gi.Trace) {
		panic("Tried to update element out of bounds")
	}

	lock(gi.lock)
	defer unlock(gi.lock)
	gi.Trace[index] = elem
}

/*
 * Get the current routine
 * Return:
 * 	the current routine
 */
func currentGoRoutine() *DedegoRoutine {
	return getg().goInfo
}

/*
 * Get the id of the current routine
 * Return:
 * 	id of the current routine, 0 if current routine is nil
 */
func GetRoutineId() uint64 {
	if currentGoRoutine() == nil {
		return 0
	}
	return currentGoRoutine().id
}

// DEDEGO-FILE-END
