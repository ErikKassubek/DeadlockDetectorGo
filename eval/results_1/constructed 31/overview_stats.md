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
| Number of routines | 19 |
| Number of spawns | 5 |
| Number of atomics | 1 |
| Number of atomic operations | 4 |
| Number of channels | 1 |
| Number of channel operations | 4 |
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
| Time for run without ADVOCATE | 0.102006 s |
| Time for run with ADVOCATE | 0.112725 s |
| Overhead of ADVOCATE | 10.508205 % |
| Replay without changes | 0.118950 s |
| Overhead of Replay | 16.610788 % s |
| Analysis | 0.043668 s |


## Results
==================== Summary ====================

-------------------- Critical -------------------
1 Potential leak with possible partner:
	channel: /home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:642@40	partner: /home/erikkassubek/Uni/HiWi/ADVOCATE/examples/constructed/potentialBugs.go:637@27
