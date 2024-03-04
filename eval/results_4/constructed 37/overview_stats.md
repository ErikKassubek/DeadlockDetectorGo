# overview Stats

## Program
| Info | Value |
| - | - |
| Number of go files | 1 |
| Number of lines | 942 |
| Number of non-empty lines | 678 |


## Trace
| Info | Value |
| - | - |
| Number of routines | 17 |
| Number of spawns | 3 |
| Number of atomics | 1 |
| Number of atomic operations | 4 |
| Number of channels | 1 |
| Number of channel operations | 2 |
| Number of selects | 0 |
| Number of select cases | 0 |
| Number of select channel operations | 0 |
| Number of select default operations | 0 |
| Number of mutexes | 1 |
| Number of mutex operations | 4 |
| Number of wait groups | 0 |
| Number of wait group operations | 0 |
| Number of cond vars | 0 |
| Number of cond var operations | 0 |
| Number of once | 0| 
| Number of once operations | 0 |


## Times
| Info | Value |
| - | - |
| Time for run without ADVOCATE | 0.101612 s |
| Time for run with ADVOCATE | 0.116048 s |
| Overhead of ADVOCATE | 14.206983 % |
| Replay without changes | 0.117832 s |
| Overhead of Replay | 15.962682 % s |
| Analysis | 0.039772 s |


## Results
==================== Summary ====================

-------------------- Critical -------------------
1 Potential mixed deadlock:
	locks: 
		/home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:741@27
		/home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:747@35
	send/close-recv: 
		/home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:742@30
		/home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:748@38