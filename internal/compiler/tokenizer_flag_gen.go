package compiler

import (
	flagrt "flag-lang/runtime"
)

// Source namespace: compiler
type SourceToken struct {
	Token  string `flag:"token"`
	Line   int64  `flag:"line"`
	Offset int64  `flag:"offset"`
}

type TokenState struct {
	Token       string `flag:"token"`
	StartLine   int64  `flag:"start-line"`
	StartOffset int64  `flag:"start-offset"`
	InString    bool   `flag:"in-string"`
	Triple      bool   `flag:"triple"`
	Escaped     bool   `flag:"escaped"`
}

type ParseToken struct {
	Kind    int64  `flag:"kind"`
	Lexeme  string `flag:"lexeme"`
	String  string `flag:"string"`
	Message string `flag:"message"`
	Line    int64  `flag:"line"`
	Col     int64  `flag:"col"`
}

func stdlib__second_arity_1(coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return flagrt.First(flagrt.Rest(coll))
	}()
}

func stdlib__second_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("stdlib__second expects exactly 1 arguments")
	}
	return stdlib__second_arity_1(args[0])
}

func stdlib__third_arity_1(coll flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return flagrt.First(flagrt.Rest(flagrt.Rest(coll)))
	}()
}

func stdlib__third_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("stdlib__third expects exactly 1 arguments")
	}
	return stdlib__third_arity_1(args[0])
}

func compiler__whitespace_char_q_arity_1(ch flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return flagrt.NewBool(flagrt.Contains(compiler__whitespace_chars, ch))
	}()
}

func compiler__whitespace_char_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__whitespace_char_q expects exactly 1 arguments")
	}
	return compiler__whitespace_char_q_arity_1(args[0])
}

func compiler__delimiter_char_q_arity_1(ch flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return flagrt.NewBool(flagrt.Contains(compiler__delimiter_chars, ch))
	}()
}

func compiler__delimiter_char_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__delimiter_char_q expects exactly 1 arguments")
	}
	return compiler__delimiter_char_q_arity_1(args[0])
}

func compiler____gtSourceToken_arity_3(token flagrt.Value, line flagrt.Value, offset flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(SourceToken{Token: token.StringValue(), Line: line.Long(), Offset: offset.Long()})
}

func compiler____gtSourceToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler____gtSourceToken expects exactly 3 arguments")
	}
	return compiler____gtSourceToken_arity_3(args[0], args[1], args[2])
}

func compiler__map__gtSourceToken_arity_1(m flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(SourceToken{Token: flagrt.Get(m, flagKw_token).StringValue(), Line: flagrt.Get(m, flagKw_line).Long(), Offset: flagrt.Get(m, flagKw_offset).Long()})
}

func compiler__map__gtSourceToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map__gtSourceToken expects exactly 1 arguments")
	}
	return compiler__map__gtSourceToken_arity_1(args[0])
}

func compiler____gtTokenState_arity_6(token flagrt.Value, start_line flagrt.Value, start_offset flagrt.Value, in_string flagrt.Value, triple flagrt.Value, escaped flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(TokenState{Token: token.StringValue(), StartLine: start_line.Long(), StartOffset: start_offset.Long(), InString: in_string.Bool(), Triple: triple.Bool(), Escaped: escaped.Bool()})
}

func compiler____gtTokenState_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 6 {
		panic("compiler____gtTokenState expects exactly 6 arguments")
	}
	return compiler____gtTokenState_arity_6(args[0], args[1], args[2], args[3], args[4], args[5])
}

func compiler__map__gtTokenState_arity_1(m flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(TokenState{Token: flagrt.Get(m, flagKw_token).StringValue(), StartLine: flagrt.Get(m, flagKw_start_line).Long(), StartOffset: flagrt.Get(m, flagKw_start_offset).Long(), InString: flagrt.Get(m, flagKw_in_string).Bool(), Triple: flagrt.Get(m, flagKw_triple).Bool(), Escaped: flagrt.Get(m, flagKw_escaped).Bool()})
}

func compiler__map__gtTokenState_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map__gtTokenState expects exactly 1 arguments")
	}
	return compiler__map__gtTokenState_arity_1(args[0])
}

func compiler____gtParseToken_arity_6(kind flagrt.Value, lexeme flagrt.Value, string flagrt.Value, message flagrt.Value, line flagrt.Value, col flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(ParseToken{Kind: kind.Long(), Lexeme: lexeme.StringValue(), String: string.StringValue(), Message: message.StringValue(), Line: line.Long(), Col: col.Long()})
}

func compiler____gtParseToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 6 {
		panic("compiler____gtParseToken expects exactly 6 arguments")
	}
	return compiler____gtParseToken_arity_6(args[0], args[1], args[2], args[3], args[4], args[5])
}

func compiler__map__gtParseToken_arity_1(m flagrt.Value) flagrt.Value {
	return flagrt.NewRecord(ParseToken{Kind: flagrt.Get(m, flagKw_kind).Long(), Lexeme: flagrt.Get(m, flagKw_lexeme).StringValue(), String: flagrt.Get(m, flagKw_string).StringValue(), Message: flagrt.Get(m, flagKw_message).StringValue(), Line: flagrt.Get(m, flagKw_line).Long(), Col: flagrt.Get(m, flagKw_col).Long()})
}

func compiler__map__gtParseToken_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__map__gtParseToken expects exactly 1 arguments")
	}
	return compiler__map__gtParseToken_arity_1(args[0])
}

func compiler__emit_token_bang_arity_4(out flagrt.Value, token flagrt.Value, line flagrt.Value, offset flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(token))) {
				return flagrt.NilValue()
			}
			return flagrt.Call(async__channel_send, out, flagrt.Call(compiler____gtSourceToken, token, line, offset))
		}()
	}()
}

func compiler__emit_token_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 4 {
		panic("compiler__emit_token_bang expects exactly 4 arguments")
	}
	return compiler__emit_token_bang_arity_4(args[0], args[1], args[2], args[3])
}

func compiler__make_state_arity_6(token flagrt.Value, start_line flagrt.Value, start_offset flagrt.Value, in_string flagrt.Value, triple flagrt.Value, escaped flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return flagrt.Call(compiler____gtTokenState, token, start_line, start_offset, in_string, triple, escaped)
	}()
}

func compiler__make_state_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 6 {
		panic("compiler__make_state expects exactly 6 arguments")
	}
	return compiler__make_state_arity_6(args[0], args[1], args[2], args[3], args[4], args[5])
}

func compiler__starts_triple_quote_q_arity_1(chars flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars))))) {
				return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars)))
			}
			if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Rest(chars)))))) {
				return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Rest(chars))))
			}
			return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Call(stdlib__third, chars)))
		}()
	}()
}

func compiler__starts_triple_quote_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__starts_triple_quote_q expects exactly 1 arguments")
	}
	return compiler__starts_triple_quote_q_arity_1(args[0])
}

func compiler__starts_dispatch_set_q_arity_1(chars flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.First(chars))))) {
				return flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.First(chars)))
			}
			return flagrt.NewBool(flagrt.Eq(flagrt.NewString("{"), flagrt.First(flagrt.Rest(chars))))
		}()
	}()
}

func compiler__starts_dispatch_set_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__starts_dispatch_set_q expects exactly 1 arguments")
	}
	return compiler__starts_dispatch_set_q_arity_1(args[0])
}

func compiler__starts_dispatch_fn_q_arity_1(chars flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.First(chars))))) {
				return flagrt.NewBool(flagrt.Eq(flagrt.NewString("#"), flagrt.First(chars)))
			}
			return flagrt.NewBool(flagrt.Eq(flagrt.NewString("("), flagrt.First(flagrt.Rest(chars))))
		}()
	}()
}

func compiler__starts_dispatch_fn_q_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__starts_dispatch_fn_q expects exactly 1 arguments")
	}
	return compiler__starts_dispatch_fn_q_arity_1(args[0])
}

func compiler__strip_quoted_token_arity_3(token flagrt.Value, prefix flagrt.Value, suffix flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var chars = flagrt.Seq(token)
			var total = flagrt.NewLong(int64(flagrt.Count(chars)))
			return func() flagrt.Value {
				if flagrt.Le(total, flagrt.Add(prefix, suffix)) {
					return flagStr_
				}
				return func() flagrt.Value {
					var remaining = flagrt.Take(flagrt.Sub(flagrt.Sub(total, prefix), suffix), flagrt.Drop(prefix, chars))
					var out = flagStr_
					for {
						__loopResult := func() flagrt.Value {
							return func() flagrt.Value {
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
									return out
								}
								return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.NewString(flagrt.Str(out, flagrt.First(remaining))))
							}()
						}()
						if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
							if len(__recurValues) != 2 {
								panic("internal error: recur arity mismatch")
							}
							remaining = __recurValues[0]
							out = __recurValues[1]
							continue
						}
						return __loopResult
					}
				}()
			}()
		}()
	}()
}

func compiler__strip_quoted_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__strip_quoted_token expects exactly 3 arguments")
	}
	return compiler__strip_quoted_token_arity_3(args[0], args[1], args[2])
}

func compiler__token__gtstring_value_arity_1(token flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var chars = flagrt.Seq(token)
			var n = flagrt.NewLong(int64(flagrt.Count(chars)))
			return func() flagrt.Value {
				if flagrt.IsTruthy(func() flagrt.Value {
					if !(flagrt.Ge(n, flagrt.NewLong(6))) {
						return flagrt.NewBool(flagrt.Ge(n, flagrt.NewLong(6)))
					}
					if !(flagrt.IsTruthy(flagrt.Call(compiler__starts_triple_quote_q, chars))) {
						return flagrt.Call(compiler__starts_triple_quote_q, chars)
					}
					if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(3)), chars)))))) {
						return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(3)), chars))))
					}
					if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(2)), chars)))))) {
						return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(2)), chars))))
					}
					return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(1)), chars))))
				}()) {
					return flagrt.Call(compiler__strip_quoted_token, token, flagrt.NewLong(3), flagrt.NewLong(3))
				}
				return func() flagrt.Value {
					if flagrt.IsTruthy(func() flagrt.Value {
						if !(flagrt.Ge(n, flagrt.NewLong(2))) {
							return flagrt.NewBool(flagrt.Ge(n, flagrt.NewLong(2)))
						}
						if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars))))) {
							return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(chars)))
						}
						return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Drop(flagrt.Sub(n, flagrt.NewLong(1)), chars))))
					}()) {
						return flagrt.Call(compiler__strip_quoted_token, token, flagrt.NewLong(1), flagrt.NewLong(1))
					}
					return func() flagrt.Value {
						if flagrt.IsTruthy(flagrt.NewBool(true)) {
							return token
						}
						return flagrt.NilValue()
					}()
				}()
			}()
		}()
	}()
}

func compiler__token__gtstring_value_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__token__gtstring_value expects exactly 1 arguments")
	}
	return compiler__token__gtstring_value_arity_1(args[0])
}

func compiler__source_token__gtparse_token_arity_1(__arg0 flagrt.Value) flagrt.Value {
	var token = flagrt.Get(__arg0, flagKw_token)
	_ = token
	var line = flagrt.Get(__arg0, flagKw_line)
	_ = line
	var offset = flagrt.Get(__arg0, flagKw_offset)
	_ = offset
	return func() flagrt.Value {
		return func() flagrt.Value {
			if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("(")))) {
				return flagrt.Call(compiler____gtParseToken, compiler__token_list_open, flagStr_, flagStr_, flagStr_, line, offset)
			}
			return func() flagrt.Value {
				if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString(")")))) {
					return flagrt.Call(compiler____gtParseToken, compiler__token_list_close, flagStr_, flagStr_, flagStr_, line, offset)
				}
				return func() flagrt.Value {
					if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("[")))) {
						return flagrt.Call(compiler____gtParseToken, compiler__token_vector_open, flagStr_, flagStr_, flagStr_, line, offset)
					}
					return func() flagrt.Value {
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("]")))) {
							return flagrt.Call(compiler____gtParseToken, compiler__token_vector_close, flagStr_, flagStr_, flagStr_, line, offset)
						}
						return func() flagrt.Value {
							if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("{")))) {
								return flagrt.Call(compiler____gtParseToken, compiler__token_map_open, flagStr_, flagStr_, flagStr_, line, offset)
							}
							return func() flagrt.Value {
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("}")))) {
									return flagrt.Call(compiler____gtParseToken, compiler__token_map_close, flagStr_, flagStr_, flagStr_, line, offset)
								}
								return func() flagrt.Value {
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("^")))) {
										return flagrt.Call(compiler____gtParseToken, compiler__token_metadata, flagStr_, flagStr_, flagStr_, line, offset)
									}
									return func() flagrt.Value {
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("'")))) {
											return flagrt.Call(compiler____gtParseToken, compiler__token_quote, flagStr_, flagStr_, flagStr_, line, offset)
										}
										return func() flagrt.Value {
											if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#{")))) {
												return flagrt.Call(compiler____gtParseToken, compiler__token_dispatch_set_open, flagStr_, flagStr_, flagStr_, line, offset)
											}
											return func() flagrt.Value {
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#(")))) {
													return flagrt.Call(compiler____gtParseToken, compiler__token_dispatch_fn_open, flagStr_, flagStr_, flagStr_, line, offset)
												}
												return func() flagrt.Value {
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(token, flagrt.NewString("#")))) {
														return flagrt.Call(compiler____gtParseToken, compiler__token_error, flagStr_, flagStr_, flagStr_unexpected_end_after__, line, offset)
													}
													return func() flagrt.Value {
														if flagrt.IsTruthy(func() flagrt.Value {
															if !(flagrt.Ge(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(2))) {
																return flagrt.NewBool(flagrt.Ge(flagrt.NewLong(int64(flagrt.Count(token))), flagrt.NewLong(2)))
															}
															if !(flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Seq(token)))))) {
																return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.First(flagrt.Seq(token))))
															}
															return flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), flagrt.Last(flagrt.Seq(token))))
														}()) {
															return flagrt.Call(compiler____gtParseToken, compiler__token_string, flagStr_, flagrt.Call(compiler__token__gtstring_value, token), flagStr_, line, offset)
														}
														return func() flagrt.Value {
															if flagrt.IsTruthy(flagrt.NewBool(true)) {
																return flagrt.Call(compiler____gtParseToken, compiler__token_atom, token, flagStr_, flagStr_, line, offset)
															}
															return flagrt.NilValue()
														}()
													}()
												}()
											}()
										}()
									}()
								}()
							}()
						}()
					}()
				}()
			}()
		}()
	}()
}

func compiler__source_token__gtparse_token_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__source_token__gtparse_token expects exactly 1 arguments")
	}
	return compiler__source_token__gtparse_token_arity_1(args[0])
}

func compiler__token_end_col_arity_1(source_token flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return flagrt.Add(flagrt.Call(flagKw_offset, source_token), flagrt.NewLong(int64(flagrt.Count(flagrt.Call(flagKw_token, source_token)))))
	}()
}

func compiler__token_end_col_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__token_end_col expects exactly 1 arguments")
	}
	return compiler__token_end_col_arity_1(args[0])
}

func compiler__tokenize_line_step_bang_arity_5(out flagrt.Value, chars flagrt.Value, line flagrt.Value, column flagrt.Value, state flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var remaining = chars
			var col = column
			var current = state
			for {
				__loopResult := func() flagrt.Value {
					return func() flagrt.Value {
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
							return func() flagrt.Value {
								if flagrt.IsTruthy(flagrt.Call(flagKw_in_string, current)) {
									return current
								}
								return func() flagrt.Value {
									_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
									return compiler__empty_token_state
								}()
							}()
						}
						return func() flagrt.Value {
							var ch = flagrt.First(remaining)
							return func() flagrt.Value {
								if flagrt.IsTruthy(flagrt.Call(flagKw_in_string, current)) {
									return func() flagrt.Value {
										if flagrt.IsTruthy(flagrt.Call(flagKw_triple, current)) {
											return func() flagrt.Value {
												if flagrt.IsTruthy(flagrt.Call(compiler__starts_triple_quote_q, remaining)) {
													return func() flagrt.Value {
														var token = flagrt.Str(flagrt.Call(flagKw_token, current), "\"\"\"")
														_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
														return flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(3), remaining), flagrt.Add(col, flagrt.NewLong(3)), compiler__empty_token_state)
													}()
												}
												return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_token, current), ch)), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(true), flagrt.NewBool(false)))
											}()
										}
										return func() flagrt.Value {
											var token = flagrt.Str(flagrt.Call(flagKw_token, current), ch)
											return func() flagrt.Value {
												if flagrt.IsTruthy(flagrt.Call(flagKw_escaped, current)) {
													return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(false)))
												}
												return func() flagrt.Value {
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\\"), ch))) {
														return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(true)))
													}
													return func() flagrt.Value {
														if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), ch))) {
															return func() flagrt.Value {
																_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
																return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), compiler__empty_token_state)
															}()
														}
														return func() flagrt.Value {
															if flagrt.IsTruthy(flagrt.NewBool(true)) {
																return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, flagrt.NewString(token), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(false)))
															}
															return flagrt.NilValue()
														}()
													}()
												}()
											}()
										}()
									}()
								}
								return func() flagrt.Value {
									if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString(";"), ch))) {
										return func() flagrt.Value {
											_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
											return compiler__empty_token_state
										}()
									}
									return func() flagrt.Value {
										if flagrt.IsTruthy(flagrt.Call(compiler__whitespace_char_q, ch)) {
											return func() flagrt.Value {
												_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
												return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), compiler__empty_token_state)
											}()
										}
										return func() flagrt.Value {
											if flagrt.IsTruthy(flagrt.Call(compiler__delimiter_char_q, ch)) {
												return func() flagrt.Value {
													_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
													_ = flagrt.Call(compiler__emit_token_bang, out, ch, line, col)
													return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), compiler__empty_token_state)
												}()
											}
											return func() flagrt.Value {
												if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\""), ch))) {
													return func() flagrt.Value {
														if flagrt.IsTruthy(flagrt.Call(compiler__starts_triple_quote_q, remaining)) {
															return func() flagrt.Value {
																_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
																return flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(3), remaining), flagrt.Add(col, flagrt.NewLong(3)), flagrt.Call(compiler__make_state, flagStr____, line, col, flagrt.NewBool(true), flagrt.NewBool(true), flagrt.NewBool(false)))
															}()
														}
														return func() flagrt.Value {
															_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
															return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, flagStr__, line, col, flagrt.NewBool(true), flagrt.NewBool(false), flagrt.NewBool(false)))
														}()
													}()
												}
												return func() flagrt.Value {
													if flagrt.IsTruthy(flagrt.Call(compiler__starts_dispatch_set_q, remaining)) {
														return func() flagrt.Value {
															_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
															_ = flagrt.Call(compiler__emit_token_bang, out, flagStr___, line, col)
															return flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(2), remaining), flagrt.Add(col, flagrt.NewLong(2)), compiler__empty_token_state)
														}()
													}
													return func() flagrt.Value {
														if flagrt.IsTruthy(flagrt.Call(compiler__starts_dispatch_fn_q, remaining)) {
															return func() flagrt.Value {
																_ = flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, current), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current))
																_ = flagrt.Call(compiler__emit_token_bang, out, flagStr____1, line, col)
																return flagrt.NewRecur(flagrt.Drop(flagrt.NewLong(2), remaining), flagrt.Add(col, flagrt.NewLong(2)), compiler__empty_token_state)
															}()
														}
														return func() flagrt.Value {
															if flagrt.IsTruthy(flagrt.NewBool(true)) {
																return func() flagrt.Value {
																	if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Call(flagKw_token, current)))) {
																		return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, ch, line, col, flagrt.NewBool(false), flagrt.NewBool(false), flagrt.NewBool(false)))
																	}
																	return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(col, flagrt.NewLong(1)), flagrt.Call(compiler__make_state, flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_token, current), ch)), flagrt.Call(flagKw_start_line, current), flagrt.Call(flagKw_start_offset, current), flagrt.NewBool(false), flagrt.NewBool(false), flagrt.NewBool(false)))
																}()
															}
															return flagrt.NilValue()
														}()
													}()
												}()
											}()
										}()
									}()
								}()
							}()
						}()
					}()
				}()
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 3 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					col = __recurValues[1]
					current = __recurValues[2]
					continue
				}
				return __loopResult
			}
		}()
	}()
}

func compiler__tokenize_line_step_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 5 {
		panic("compiler__tokenize_line_step_bang expects exactly 5 arguments")
	}
	return compiler__tokenize_line_step_bang_arity_5(args[0], args[1], args[2], args[3], args[4])
}

func compiler__tokenize_lines_bang_arity_3(out flagrt.Value, lines flagrt.Value, line_number flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var remaining = lines
			var line = line_number
			var state = compiler__empty_token_state
			for {
				__loopResult := func() flagrt.Value {
					return func() flagrt.Value {
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(remaining))) {
							return flagrt.Call(compiler__emit_token_bang, out, flagrt.Call(flagKw_token, state), flagrt.Call(flagKw_start_line, state), flagrt.Call(flagKw_start_offset, state))
						}
						return func() flagrt.Value {
							var next_state = flagrt.Call(compiler__tokenize_line_step_bang, out, flagrt.Seq(flagrt.NewString(flagrt.Str(flagrt.First(remaining)))), line, flagrt.NewLong(1), state)
							var has_more = flagrt.NewBool(!flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(flagrt.Rest(remaining)))))
							var carry_state = func() flagrt.Value {
								if flagrt.IsTruthy(func() flagrt.Value {
									if !(flagrt.IsTruthy(has_more)) {
										return has_more
									}
									return flagrt.Call(flagKw_in_string, next_state)
								}()) {
									return flagrt.Call(compiler__make_state, flagrt.NewString(flagrt.Str(flagrt.Call(flagKw_token, next_state), "\n")), flagrt.Call(flagKw_start_line, next_state), flagrt.Call(flagKw_start_offset, next_state), flagrt.NewBool(true), flagrt.Call(flagKw_triple, next_state), flagrt.Call(flagKw_escaped, next_state))
								}
								return next_state
							}()
							return flagrt.NewRecur(flagrt.Rest(remaining), flagrt.Add(line, flagrt.NewLong(1)), carry_state)
						}()
					}()
				}()
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 3 {
						panic("internal error: recur arity mismatch")
					}
					remaining = __recurValues[0]
					line = __recurValues[1]
					state = __recurValues[2]
					continue
				}
				return __loopResult
			}
		}()
	}()
}

func compiler__tokenize_lines_bang_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 3 {
		panic("compiler__tokenize_lines_bang expects exactly 3 arguments")
	}
	return compiler__tokenize_lines_bang_arity_3(args[0], args[1], args[2])
}

func compiler__split_lines_arity_1(source flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var chars = flagrt.Seq(source)
			var line = flagStr_
			var lines = flagVec
			for {
				__loopResult := func() flagrt.Value {
					return func() flagrt.Value {
						if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(chars))) {
							return func() flagrt.Value {
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsEmpty(line))) {
									return lines
								}
								return flagrt.Conj(lines, line)
							}()
						}
						return func() flagrt.Value {
							var ch = flagrt.First(chars)
							return func() flagrt.Value {
								if flagrt.IsTruthy(flagrt.NewBool(flagrt.Eq(flagrt.NewString("\n"), ch))) {
									return flagrt.NewRecur(flagrt.Rest(chars), flagStr_, flagrt.Conj(lines, line))
								}
								return flagrt.NewRecur(flagrt.Rest(chars), flagrt.NewString(flagrt.Str(line, ch)), lines)
							}()
						}()
					}()
				}()
				if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
					if len(__recurValues) != 3 {
						panic("internal error: recur arity mismatch")
					}
					chars = __recurValues[0]
					line = __recurValues[1]
					lines = __recurValues[2]
					continue
				}
				return __loopResult
			}
		}()
	}()
}

func compiler__split_lines_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__split_lines expects exactly 1 arguments")
	}
	return compiler__split_lines_arity_1(args[0])
}

// Tokenize an in-memory source string and return a channel of SourceToken maps {:token :line :offset}.
func compiler__tokenize_source_arity_1(source flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
			_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 0 {
					panic("fn expects exactly 0 arguments")
				}
				return func() flagrt.Value {
					return func() flagrt.Value {
						_ = flagrt.Call(compiler__tokenize_lines_bang, out, flagrt.Call(compiler__split_lines, source), flagrt.NewLong(1))
						return flagrt.Call(async__channel_close, out)
					}()
				}()
			}))
			return out
		}()
	}()
}

func compiler__tokenize_source_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__tokenize_source expects exactly 1 arguments")
	}
	return compiler__tokenize_source_arity_1(args[0])
}

// Read a source file and return a channel of token maps {:token :line :offset}.
func compiler__tokenize_file_arity_1(path flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
			_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 0 {
					panic("fn expects exactly 0 arguments")
				}
				return func() flagrt.Value {
					return func() flagrt.Value {
						var __bind0 = flagrt.Call(flagrt.GoBind_runtime_OpenFile, path)
						defer __bind0.Close()
						var rdr = __bind0
						return func() flagrt.Value {
							_ = flagrt.Call(compiler__tokenize_lines_bang, out, flagrt.LineSeq(rdr), flagrt.NewLong(1))
							return flagrt.Call(async__channel_close, out)
						}()
					}()
				}()
			}))
			return out
		}()
	}()
}

func compiler__tokenize_file_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__tokenize_file expects exactly 1 arguments")
	}
	return compiler__tokenize_file_arity_1(args[0])
}

// Read a source file and return a channel of ParseToken-shaped maps matching Go ParseToken fields.
func compiler__tokenize_file_parse_tokens_arity_1(path flagrt.Value) flagrt.Value {
	return func() flagrt.Value {
		return func() flagrt.Value {
			var out = flagrt.Call(async__make_channel, flagrt.NewLong(64))
			var in = flagrt.Call(compiler__tokenize_file, path)
			_ = flagrt.Call(async__go_run, flagrt.NewFunction(func(args ...flagrt.Value) flagrt.Value {
				if len(args) != 0 {
					panic("fn expects exactly 0 arguments")
				}
				return func() flagrt.Value {
					return func() flagrt.Value {
						var last_token = flagrt.NilValue()
						for {
							__loopResult := func() flagrt.Value {
								return func() flagrt.Value {
									var source_token = flagrt.Call(async__channel_receive, in)
									return func() flagrt.Value {
										if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(source_token))) {
											return func() flagrt.Value {
												var eof_line = func() flagrt.Value {
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(last_token))) {
														return flagrt.NewLong(1)
													}
													return flagrt.Call(flagKw_line, last_token)
												}()
												var eof_col = func() flagrt.Value {
													if flagrt.IsTruthy(flagrt.NewBool(flagrt.IsNil(last_token))) {
														return flagrt.NewLong(1)
													}
													return flagrt.Call(compiler__token_end_col, last_token)
												}()
												_ = flagrt.Call(async__channel_send, out, flagrt.Call(compiler____gtParseToken, compiler__token_eof, flagStr_, flagStr_, flagStr_, eof_line, eof_col))
												return flagrt.Call(async__channel_close, out)
											}()
										}
										return func() flagrt.Value {
											_ = flagrt.Call(async__channel_send, out, flagrt.Call(compiler__source_token__gtparse_token, source_token))
											return flagrt.NewRecur(source_token)
										}()
									}()
								}()
							}()
							if __recurValues, __isRecur := flagrt.UnwrapRecur(__loopResult); __isRecur {
								if len(__recurValues) != 1 {
									panic("internal error: recur arity mismatch")
								}
								last_token = __recurValues[0]
								continue
							}
							return __loopResult
						}
					}()
				}()
			}))
			return out
		}()
	}()
}

func compiler__tokenize_file_parse_tokens_variadic(args ...flagrt.Value) flagrt.Value {
	if len(args) != 1 {
		panic("compiler__tokenize_file_parse_tokens expects exactly 1 arguments")
	}
	return compiler__tokenize_file_parse_tokens_arity_1(args[0])
}

var async__channel_lines = flagrt.GoBind_async_LinesPipe
var async__go_run = flagrt.GoBind_async_GoRun
var async__sleep = flagrt.GoBind_async_Sleep
var async__select_ = flagrt.GoBind_async_Select
var async__channel_map = flagrt.GoBind_async_PipeMap
var async__make_channel = flagrt.GoBind_async_MakeChannel
var async__channel_close = flagrt.GoBind_async_ChannelClose
var async__channel_some_q = flagrt.GoBind_async_PipeSome
var async__channel_send = flagrt.GoBind_async_ChannelSend
var async__future_run = flagrt.GoBind_async_FutureRun
var async__future_piped_run = flagrt.GoBind_async_FuturePipeRun
var async__channel_receive = flagrt.GoBind_async_ChannelReceive
var async__channel_filter = flagrt.GoBind_async_PipeFilter
var async__channel_reduce = flagrt.GoBind_async_PipeReduce
var async__channel_every_q = flagrt.GoBind_async_PipeEvery
var stdlib__second = flagrt.NewFunction(stdlib__second_variadic)
var stdlib__third = flagrt.NewFunction(stdlib__third_variadic)
var compiler__whitespace_chars = flagSet
var compiler__delimiter_chars = flagSet_1
var compiler__token_eof = flagrt.NewLong(0)
var compiler__token_error = flagrt.NewLong(1)
var compiler__token_list_open = flagrt.NewLong(2)
var compiler__token_list_close = flagrt.NewLong(3)
var compiler__token_vector_open = flagrt.NewLong(4)
var compiler__token_vector_close = flagrt.NewLong(5)
var compiler__token_map_open = flagrt.NewLong(6)
var compiler__token_map_close = flagrt.NewLong(7)
var compiler__token_metadata = flagrt.NewLong(8)
var compiler__token_quote = flagrt.NewLong(9)
var compiler__token_dispatch_set_open = flagrt.NewLong(10)
var compiler__token_dispatch_fn_open = flagrt.NewLong(11)
var compiler__token_string = flagrt.NewLong(12)
var compiler__token_atom = flagrt.NewLong(13)
var compiler__whitespace_char_q = flagrt.NewFunction(compiler__whitespace_char_q_variadic)
var compiler__delimiter_char_q = flagrt.NewFunction(compiler__delimiter_char_q_variadic)
var compiler____gtSourceToken = flagrt.NewFunction(compiler____gtSourceToken_variadic)
var compiler__map__gtSourceToken = flagrt.NewFunction(compiler__map__gtSourceToken_variadic)
var compiler____gtTokenState = flagrt.NewFunction(compiler____gtTokenState_variadic)
var compiler__map__gtTokenState = flagrt.NewFunction(compiler__map__gtTokenState_variadic)
var compiler____gtParseToken = flagrt.NewFunction(compiler____gtParseToken_variadic)
var compiler__map__gtParseToken = flagrt.NewFunction(compiler__map__gtParseToken_variadic)
var compiler__emit_token_bang = flagrt.NewFunction(compiler__emit_token_bang_variadic)
var compiler__make_state = flagrt.NewFunction(compiler__make_state_variadic)
var compiler__empty_token_state = flagrt.Call(compiler__make_state, flagStr_, flagrt.NewLong(0), flagrt.NewLong(0), flagrt.NewBool(false), flagrt.NewBool(false), flagrt.NewBool(false))
var compiler__starts_triple_quote_q = flagrt.NewFunction(compiler__starts_triple_quote_q_variadic)
var compiler__starts_dispatch_set_q = flagrt.NewFunction(compiler__starts_dispatch_set_q_variadic)
var compiler__starts_dispatch_fn_q = flagrt.NewFunction(compiler__starts_dispatch_fn_q_variadic)
var compiler__strip_quoted_token = flagrt.NewFunction(compiler__strip_quoted_token_variadic)
var compiler__token__gtstring_value = flagrt.NewFunction(compiler__token__gtstring_value_variadic)
var compiler__source_token__gtparse_token = flagrt.NewFunction(compiler__source_token__gtparse_token_variadic)
var compiler__token_end_col = flagrt.NewFunction(compiler__token_end_col_variadic)
var compiler__tokenize_line_step_bang = flagrt.NewFunction(compiler__tokenize_line_step_bang_variadic)
var compiler__tokenize_lines_bang = flagrt.NewFunction(compiler__tokenize_lines_bang_variadic)
var compiler__split_lines = flagrt.NewFunction(compiler__split_lines_variadic)
var compiler__tokenize_source = flagrt.NewFunction(compiler__tokenize_source_variadic)
var compiler__tokenize_file = flagrt.NewFunction(compiler__tokenize_file_variadic)
var compiler__tokenize_file_parse_tokens = flagrt.NewFunction(compiler__tokenize_file_parse_tokens_variadic)
var flagSet = flagrt.NewSet(flagrt.NewString(" "), flagrt.NewString("\t"), flagrt.NewString("\n"), flagrt.NewString("\r"), flagrt.NewString(","))
var flagSet_1 = flagrt.NewSet(flagrt.NewString("("), flagrt.NewString(")"), flagrt.NewString("["), flagrt.NewString("]"), flagrt.NewString("{"), flagrt.NewString("}"), flagrt.NewString("^"), flagrt.NewString("'"))
var flagKw_token = flagrt.NewKeyword("token")
var flagKw_line = flagrt.NewKeyword("line")
var flagKw_offset = flagrt.NewKeyword("offset")
var flagKw_start_line = flagrt.NewKeyword("start-line")
var flagKw_start_offset = flagrt.NewKeyword("start-offset")
var flagKw_in_string = flagrt.NewKeyword("in-string")
var flagKw_triple = flagrt.NewKeyword("triple")
var flagKw_escaped = flagrt.NewKeyword("escaped")
var flagKw_kind = flagrt.NewKeyword("kind")
var flagKw_lexeme = flagrt.NewKeyword("lexeme")
var flagKw_string = flagrt.NewKeyword("string")
var flagKw_message = flagrt.NewKeyword("message")
var flagKw_col = flagrt.NewKeyword("col")
var flagStr_ = flagrt.NewString("")
var flagStr_unexpected_end_after__ = flagrt.NewString("unexpected end after #")
var flagStr____ = flagrt.NewString("\"\"\"")
var flagStr__ = flagrt.NewString("\"")
var flagStr___ = flagrt.NewString("#{")
var flagStr____1 = flagrt.NewString("#(")
var flagVec = flagrt.NewArray()
