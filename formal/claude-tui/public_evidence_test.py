import unittest

from public_evidence import redact_tlc_output


class PublicEvidenceTests(unittest.TestCase):
    def test_machine_banner_does_not_publish_host_pid_or_checkout(self):
        raw = (
            "TLC2 Version 2.19\n"
            "Running breadth-first search Model-Checking with fp 0 and seed 1 with 1 worker on 8 cores [pid: 123] (Linux example-host.invalid).\n"
            "Parsing file /home/example/project/formal/claude-tui/models/Case.tla\n"
            "Parsing file /tmp/Naturals.tla\n"
            "Error: Invariant CallerIsolation is violated.\n"
            "12 states generated, 11 distinct states found\n"
        )
        public = redact_tlc_output(raw)
        self.assertNotIn("example-host.invalid", public)
        self.assertNotIn("pid:", public)
        self.assertNotIn("/home/", public)
        self.assertIn("fp 0 and seed 1 with 1 worker", public)
        self.assertIn("Parsing file formal/claude-tui/models/Case.tla\n", public)
        self.assertIn("Parsing file Naturals.tla\n", public)
        self.assertIn("Error: Invariant CallerIsolation is violated.\n", public)
        self.assertIn("12 states generated, 11 distinct states found\n", public)
        self.assertEqual(redact_tlc_output(public), public)

    def test_windows_paths_and_no_final_newline(self):
        raw = r"Parsing file C:\Users\example\repo\formal\claude-tui\models\Case.tla"
        self.assertEqual(redact_tlc_output(raw), "Parsing file formal/claude-tui/models/Case.tla")
        self.assertEqual(redact_tlc_output("No error has been found."), "No error has been found.")


if __name__ == "__main__":
    unittest.main()
