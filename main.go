package main

// #cgo pkg-config: libinput
/*
#include <errno.h>
#include <fcntl.h>
#include <libinput.h>
#include <poll.h>
#include <unistd.h>

int open_restricted(const char *path, int flags, void *user_data)
{
	int fd = open(path, flags);
	return fd < 0 ? -errno : fd;
}

void close_restricted(int fd, void *user_data)
{
	close(fd);
}

const struct libinput_interface iface = {
        .open_restricted = open_restricted,
        .close_restricted = close_restricted,
};
*/
import "C"

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"unsafe"
)

func main() {
	li := C.libinput_path_create_context(&C.iface, nil)

	devicePath := C.CString("/dev/input/event4")
	C.libinput_path_add_device(li, devicePath)

	if handleAndProcessEvents(li, nil) == 0 {
		log.Fatal("Expected device added events on startup but got none.")
	}

	// Create a channel to receive events from the main loop.
	eventChan := make(chan GestureEvent)
	finished := make(chan bool)
	go mainLoop(li, eventChan, finished)
	processingLoop(eventChan)

	// Wait for finished signal to arrive, then close the channel and clean up
	// related resources.
	<-finished
	close(finished)
	close(eventChan)

	C.libinput_unref(li)
	C.free(unsafe.Pointer(devicePath))

	fmt.Println("Hello")
}

func mainLoop(li *C.struct_libinput, eventChan chan GestureEvent, finished chan bool) {
	fds := C.struct_pollfd{
		fd:      C.libinput_get_fd(li),
		events:  C.POLLIN,
		revents: 0,
	}

	for {
		if C.poll(&fds, 1, -1) > -1 {
			handleAndProcessEvents(li, eventChan)
		}
	}

	finished <- true
}

func processingLoop(eventChan chan GestureEvent) {
	const minSamples = 3

	var currentSwipe []GestureEvent
	isCurrentSwipeProcessed := false

	for {
		event := <-eventChan

		switch event.EventType {
		case C.LIBINPUT_EVENT_GESTURE_SWIPE_BEGIN:
			currentSwipe = nil
		case C.LIBINPUT_EVENT_GESTURE_SWIPE_UPDATE:
			currentSwipe = append(currentSwipe, event)
			if len(currentSwipe) >= minSamples && !isCurrentSwipeProcessed {
				direction := getSwipeDirection(currentSwipe)
				invokeAction(event.FingerCount, direction)
				isCurrentSwipeProcessed = true
				fmt.Printf("Processing %d events, direction %s\n", len(currentSwipe), direction)
			}
		case C.LIBINPUT_EVENT_GESTURE_SWIPE_END:
			isCurrentSwipeProcessed = false
		}
	}
}

func invokeAction(numFingers int, direction SwipeDirection) {
	if numFingers == 3 && direction == SwipeDirectionTop {
		switch direction {
		case SwipeDirectionTop:
			cmd := exec.Command("xdotool", "key", "Super+W")
			cmd.Start()
		case SwipeDirectionBottom:
			cmd := exec.Command("xdotool", "key", "Super+W")
			cmd.Start()
		}
	}

	if numFingers == 4 {
		switch direction {
		case SwipeDirectionLeft:
			cmd := exec.Command("xdotool", "key", "Ctrl+Alt+Right")
			cmd.Start()
		case SwipeDirectionRight:
			cmd := exec.Command("xdotool", "key", "Ctrl+Alt+Left")
			cmd.Start()
		}
	}
}

func getSwipeDirection(swipe []GestureEvent) SwipeDirection {
	xVector, yVector := float64(0), float64(0)

	for _, sample := range swipe {
		xVector = xVector + sample.Dx
		yVector = yVector + sample.Dy
	}

	angle := math.Atan(yVector/xVector) * 180 / math.Pi
	if xVector > 0 {
		if angle > -45 && angle < 45 {
			return SwipeDirectionRight
		} else if angle < -45 {
			return SwipeDirectionTop
		} else {
			return SwipeDirectionBottom
		}
	} else {
		if angle > -45 && angle < 45 {
			return SwipeDirectionLeft
		} else if angle < -45 {
			return SwipeDirectionBottom
		} else {
			return SwipeDirectionTop
		}
	}
}

func handleAndProcessEvents(li *C.struct_libinput, eventChan chan GestureEvent) (numEvents int) {
	numEvents = 0

	C.libinput_dispatch(li)
	for {
		event := C.libinput_get_event(li)
		if event == nil {
			break
		}

		switch eventType := C.libinput_event_get_type(event); eventType {
		case C.LIBINPUT_EVENT_GESTURE_SWIPE_BEGIN:
			gestureEvent := C.libinput_event_get_gesture_event(event)
			eventChan <- GestureEvent{
				EventType:   eventType,
				FingerCount: int(C.libinput_event_gesture_get_finger_count(gestureEvent)),
			}
		case C.LIBINPUT_EVENT_GESTURE_SWIPE_UPDATE:
			gestureEvent := C.libinput_event_get_gesture_event(event)
			eventChan <- GestureEvent{
				EventType:   eventType,
				FingerCount: int(C.libinput_event_gesture_get_finger_count(gestureEvent)),
				Dx:          float64(C.libinput_event_gesture_get_dx(gestureEvent)),
				Dy:          float64(C.libinput_event_gesture_get_dy(gestureEvent)),
			}
		case C.LIBINPUT_EVENT_GESTURE_SWIPE_END:
			gestureEvent := C.libinput_event_get_gesture_event(event)
			eventChan <- GestureEvent{
				EventType:   eventType,
				FingerCount: int(C.libinput_event_gesture_get_finger_count(gestureEvent)),
			}
		}

		C.libinput_event_destroy(event)
		C.libinput_dispatch(li)

		numEvents = numEvents + 1
	}

	return numEvents
}

// GestureEvent contains information about a gesture.
type GestureEvent struct {
	EventType   uint32
	FingerCount int
	Dx          float64
	Dy          float64
}

// SwipeDirection denotes the direction of the swipe.
type SwipeDirection int

const (
	// SwipeDirectionTop denotes a swipe to the top.
	SwipeDirectionTop SwipeDirection = iota
	// SwipeDirectionRight denotes a swipe to the right.
	SwipeDirectionRight SwipeDirection = iota
	// SwipeDirectionBottom denotes a swipe to the bottom.
	SwipeDirectionBottom SwipeDirection = iota
	// SwipeDirectionLeft denotes a swipe to the left.
	SwipeDirectionLeft SwipeDirection = iota
)
