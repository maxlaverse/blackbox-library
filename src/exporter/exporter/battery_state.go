package exporter

import (
	"time"
)

const (
	adcVref = 33.0
)

type batteryState struct {
	currentAmps         float64
	energyMilliampHours float64
	voltageVolt         float64
	lastTime            int32
	currentScale        int32
	currentOffset       int32
	vbatScale           int32
}

func (b *batteryState) setLatestAmperage(value int32, newTime int32) {
	b.currentAmps = float64(value*adcVref*100/4095-b.currentOffset) * 10 / float64(b.currentScale)
	if b.lastTime != 0.0 {
		b.energyMilliampHours += b.currentAmps * float64(newTime-b.lastTime) / time.Hour.Seconds() / 1000
	}
	b.lastTime = newTime
}

func (b *batteryState) setLatestVbat(value int32) {
	// ADC is 12 bit (i.e. max 0xFFF), voltage reference is 3.3V, vbatscale is premultiplied by 100
	b.voltageVolt = float64(value*adcVref*b.vbatScale) / 4095 / 100.0
}
