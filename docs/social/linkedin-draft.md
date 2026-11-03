# LinkedIn draft — not published

An unchanged accuracy score can hide a regression.

I built AI Reliability Lab in Go to explore that problem. In its reproducible demo, two runs both pass 2 of 3 tasks—but the candidate breaks one previously passing case while fixing another.

The toolkit pairs results by model and case, checks dataset digests and coverage, and flags the broken case. It includes versioned task definitions, exact/trimmed-output graders, a provider interface, JSON/HTML reports and an optional CI regression gate.

The current demo replays scripted responses. Token counts are fixtures and timing is local execution time; I am not claiming live model quality from it. That distinction also appears in the generated reports.

Code and one-command demo: https://github.com/akshay-dhangar/ai-reliability-lab

Next question: which task-level regressions should block a release, even when the overall score improves?

#Go #BackendEngineering #AI #Evaluation
