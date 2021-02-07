package main

import "errors"

// AdviceArgs struct represents the RPC method arguments
type AdviceArgs struct {
	Topic       string `json:"topic"`
	MaybeAmount *int   `json:"amount,omitempty"`
}

// AdviceReply struct represents the RPC method return structure
type AdviceReply struct {
	AdviceList []string `json:"adviceList"`
}

// AdviceService struct represents the RPC implementation
type AdviceService struct {
	getter AdviceGetter
}

// GiveMeAdvice function is the actual business logic of the RPC method
func (a *AdviceService) GiveMeAdvice(args *AdviceArgs, reply *AdviceReply) error {

	var advices []string
	var err error

	if args.MaybeAmount == nil {
		advices, err = a.getter.GetAdvicesFor(args.Topic)
	} else if *(args.MaybeAmount) >= 0 {
		advices, err = a.getter.GetAdvicesLimitedFor(args.Topic, *(args.MaybeAmount))
	} else {
		return errors.New("Cannot accept an amount that is less than 0")
	}

	if err != nil {
		return err
	}

	reply.AdviceList = advices

	return nil
}
