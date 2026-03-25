package cron

import "errors"

var ErrCronNotInitialized = errors.New("per-user cron service not initialized")
