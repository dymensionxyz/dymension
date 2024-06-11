// Package derr = dymension err is for errors besides the google ones
//
// This file should contain ubiquitous domain specific errors which warrant their own handling on top of gerr handling
// For example, if your caller code wants to differentiate between a generic failed precondition, and a failed precondition due to
// misbehavior, you would define a misbehavior error here.
//
// It is likely that there are not many of these errors, since their usage should be ubiquitous.
package derr
