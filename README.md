# Introduction
This experiment compares the performance of writing data using io_uring, o_direct, and a control method for comparing results.
The two operations tested were downloading an Ubuntu image from a nearby repository and writing zeros to a 5 GiB sparse file.
This experiment was ran on an m5d.large instance (2 vCPU, 8 GiB, 75 GiB NVMe SSD) running kernel 5.10.

# Control
The control uses standard write operations provided by GoLang including `io.Copy` and `os.File.Write`. For the overwrite
test a call to `dd` was also used to compare results against.

## Results
|                | n | P0     | p50    | p90    | p100   |
|----------------|---|--------|--------|--------|--------|
| Download       | 100 | 6,379  | 16,490 | 22,125 | 51,587 |
| Overwrite      | 100 | 56,748 | 58,265 | 58,284 | 58,474 |
| Overwrite (dd) | 100 | 56,729 | 56,772 | 56,793 | 57,382 | 

# O_DIRECT
These tests were similar to the control tests, but used the O_DIRECT flag to write directly to disk rather than 
via any kernel caches. As a result these tests will result in slower performance.

For these tests I am using [github.com/brk0v/directio](https://github.com/brk0v/directio) to write to disk using 
`O_DIRECT`.

## Results
|                | n | P0     | p50    | p90    | p100   |
|----------------|---|--------|--------|--------|--------|
| Download       | 100 | 21,119 | 27,411 | 30,979 | 53,641 |
| Overwrite      | 100 | 76,411 | 77,555 | 77,558 | 77,560 |

# IO_URING
io_uring introduces a new set of kernel system calls which were introduced in kernel 5.1 to support asynchronous I/O. 
With io_uring userspace programs can instead submit read or write requests to a circular buffer, called a submission 
queue (SQ), which is shared between userspace and kernel space. With system calls like `write()` data must be copied
from user space into kernel space. Since the io_uring buffers are shared between kernel-space and user-space, then this
eliminates this copying of data.

The kernel will asynchronously take action and write the results to a completion queue (CQ) which a userspace process 
can read to see the results of the action. Experimental evidence suggests that writing requests to 
the submission queue to write 5GiB of data of can complete in double-digit milli-seconds time. However, this can be 
misleading since the actual write operations are performed in the background asynchronously. For this experiment we are 
measuring the time it takes to both write to the submission queue and to read all of the results of the writes in the 
completion queue.

For these tests, I am using the [github.com/iceber/iouring-go](https://github.com/iceber/iouring-go) library to interact
with io_uring.

## Results
|                | n | P0     | p50    | p90    | p100   |
|----------------|---|--------|--------|--------|--------|
| Download       | 100 | 6,225  | 16,448 | 21,289 | 43,151 |
| Overwrite      | 100 | 57,847 | 58,272 | 58,293 | 58,399 |
