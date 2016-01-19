# cubewhisperer

This is my first attempt at implementing anything remotely like speech recognition. I want to make a program that learns how you say the 18 moves on a Rubik's cube (e.g. "R", "U", "L2") and then transcribes a sequence of moves as you say them.

Right now, the program can learn 8 different words (R, U, F, B, D, L, Squared, Prime), and then transcribe them as you say them. It nails the words pretty accurately, but you have to make sure to leave ample pauses between each word. I have not tested this with any background noise; that might ruin it.
