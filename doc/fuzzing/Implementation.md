 # Implementation

## Into
- It should be possible to determine all needed to determine how interesting a run is from the trace
- Replay should be adaptable, to prefer a specified select case
- Checking if a select case is possible using the HB relation would only make sense until the program run first executes a select, where a different channel is used than in the last recording. After that, the HB relation is no longer valid and can therefore not be used to determine, if a select case has a possible partner.
- Maybe the score calculation could include information from the HB relation. E.g., a run where many not executed select cases have a possible partner, could be more interesting.

## Determine whether the run was interesting
- A run is interesting, if one of the following conditions is met. The underlying information need to be stored in a file for the following runs.
  - The run contains a new pair of channel operations
    - All pairs of channel operations (send-recv) must be stored
  - If an operation pair's execution counter changes significantly from previous order.
    - For each operation pair determine the average number of executions per run
    - Determine a run to be interesting, if the number diverges from this average
    - Determine what is ment with "Specifically, if the counter falls into a range $(2^{N-1}, 2^N]$ to which no previous counter belongs" in the paper ?????
  - If a new channel operations (creation, close, not close) is triggered for the first time
    - We must be able to identify each channel
      - Add a trace element on create -> channels can be identified by path of create and channel ID
    - We must store all channels ever created
    - We must store for all channels that have ever been created, whether they have been ever closed/not closed or both
  - If a buffered channel gets a larger maximum fullness than in all previous executions
    - For each channel we must store the maximum fullness over all runs

## Determine the score
- For the base GFuzz, we need to extract the following information from the trace:
  - CountChOpPair_i: For each pair i of send/receive, how often was it executed
  - CreateCh: How many distinct channels have been created
    - Can be determined based on channel id
  - CloseCh: Number of closed channel
    - Count close operations
  - MaxChBufFull: Maximum fullness for each buffer
    - Each buffered channel info in the trace contains the current qSize. Pass all send and get the biggest
- With those values it is possible to determine the score
- Later this should be extended based on information from the happens before

## Storage of information
- The information are stored in one file called "AdvocateFuzzing.log"
- We change the recording to also include creations of channels. From this, a channel can be uniquely identified over multiple traces based on its creation location.
- A pair of send/recv is determined by the pair of paths of the operations.
- The file storing the information contains the following information. Each block is separated by a line containing "###########"
  - Info (one line separated by ;)
    - Number of runs already performed (needed to accurately compute avg. number of communications on pair)
    - Maximum score over all runs
  - For each channel that has ever been created, we store the following line
    - fileCreate:lineCreate;closeInfo;qSize;maxQSize
      - closeInfo can be
        - a: always been closed
        - n: never been closed
        - s: in some runs it was closed, but not in all
      - qSize is the buffer size
      - maxQSize is the maximum fullness of the buffer over all operations
  - For each pair of operations we store the following line:
    - fileSend:lineSend:selCaseSend;fileRecv:lineRecv:selCaseRecv;avgNumberCom
      - selCaseSend and selCaseRecv identify the cases in a select. If the send/recv is in a select, the value is set to the number of the case. If it is not part of a select, it is set to 0.

## Select File
We adapt the selects, to prefer one of the select cases.
Therefore, we create a file to store the list of preferred case ids for each
select. The file contains one line for each select of the form

- \[file\]:\[line\];\[listOfIds\]

where the \[listOfIds\] contains the IDs separated by `,`.
If in a fuzzing run, a select is more often run that specified in the file, the selects will run without a preferred case.

## Select
The internal implementation of a select consists of 3 passes.
In the first pass it is checked, whether for one of the cases an immediate communication is possible.
If this is possible, it is executed and the select is done.
If this is not the case an intermediate step checks if there is a default case. If it is, it will be executed.
In the second pass, all cases are added into the waiting queues for the corresponding channel and the select parks.
If one of the cases is woken up by another routine that is willing to communicate, the communication is executed.
In the 3rd pass, all not executed cases will be removed from the corresponding queues.\\

The fuzzing is implemented in the following way.
If the default case has been selected as preferred it is executed before the first pass and the select is done.
For the first pass, only the preferred case will be checked. If it did not directly execute,
a wait on the corresponding channel and direction for the preferred case will be started.
Additionally, a background go routine will be started.
This routine contains a timer. While the timer is running, it will
constantly monitor, whether the waiting preferred has been executed. If this is the
case, it will terminate. If the timer finishes while the channel has not executed,
it will continue by running the complete, unmodified select implementation again.