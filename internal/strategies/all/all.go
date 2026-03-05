// Package all blank-imports all built-in strategies for convenience registration.
package all

import (
	_ "github.com/fredbi/go-mutesting/internal/strategies/argswap"
	_ "github.com/fredbi/go-mutesting/internal/strategies/arithmetic"
	_ "github.com/fredbi/go-mutesting/internal/strategies/chanops"
	_ "github.com/fredbi/go-mutesting/internal/strategies/assignment"
	_ "github.com/fredbi/go-mutesting/internal/strategies/bitwise"
	_ "github.com/fredbi/go-mutesting/internal/strategies/boolliteral"
	_ "github.com/fredbi/go-mutesting/internal/strategies/branch"
	_ "github.com/fredbi/go-mutesting/internal/strategies/condexpr"
	_ "github.com/fredbi/go-mutesting/internal/strategies/conditional"
	_ "github.com/fredbi/go-mutesting/internal/strategies/expression"
	_ "github.com/fredbi/go-mutesting/internal/strategies/incdec"
	_ "github.com/fredbi/go-mutesting/internal/strategies/lockremove"
	_ "github.com/fredbi/go-mutesting/internal/strategies/logical"
	_ "github.com/fredbi/go-mutesting/internal/strategies/loopctrl"
	_ "github.com/fredbi/go-mutesting/internal/strategies/negatives"
	_ "github.com/fredbi/go-mutesting/internal/strategies/nilguard"
	_ "github.com/fredbi/go-mutesting/internal/strategies/panicremove"
	_ "github.com/fredbi/go-mutesting/internal/strategies/returnval"
	_ "github.com/fredbi/go-mutesting/internal/strategies/slicebound"
	_ "github.com/fredbi/go-mutesting/internal/strategies/statement"
	_ "github.com/fredbi/go-mutesting/internal/strategies/stringliteral"
	_ "github.com/fredbi/go-mutesting/internal/strategies/swapbranch"
)
