# Glimpse Engineering Guidelines

## Philosophy
1. **Contract-First:** Design the interface/Protobuf first, then implement the core.
2. **Clarity > Cleverness:** Optimize for maintainability. Use existing patterns over new ones.
3. **Atomic Changes:** Make the smallest change necessary to solve the problem.
4. **Simplicity** is the prerequisite of reliability.
5. **Validate Assumptions**: Verify intent and clarify ambiguity before implementation. Engage deliberate ("System 2") thinking to solve the right problem.

## Tech Stack & Constraints

### Backend (Go 1.25+)
- **Testing:** `testify/assert` @latest + `gomock` v0.6.0.
