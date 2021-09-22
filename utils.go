package main

func calculateChange(original, new, price float32) float32 {
	return price * (original - new)
}
