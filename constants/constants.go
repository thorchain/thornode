package constants

// The emission curve targets a ~2% emission after 10 years (similar to Bitcoin).
// The secondsPerBlock directly affects emission rate, and should be updated if markedly different in production
// Day 0 Emission is ~25%, Year 1 Emission is ~20%, Year 10 Emission is ~2%
const emissionCurve = 6         // An arbitrary factor to target desired curve
const secondsPerYear = 31556952 // 365.2425 * 86400
const secondsPerBlock = 5       // Number of seconds per THORChain block

// A new pool is enabled on a cycle
const newPoolCycle = 50000 // Enable a pool every 50,000 blocks (~3 days)
