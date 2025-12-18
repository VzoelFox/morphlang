package vm

import (
	"os"
	"testing"
)

func TestFileStream(t *testing.T) {
	tmpFile := "test_stream.txt"
	defer os.Remove(tmpFile)

	// 1. Write
	inputWrite := `
    f = buka_file("test_stream.txt", "t");
    tulis(f, "Halo Dunia");
    tutup_file(f);
    `
	runVmTest(t, inputWrite, nil)

	// 2. Read All
	inputRead := `
    f = buka_file("test_stream.txt", "b");
    content = baca(f);
    tutup_file(f);
    content;
    `
	runVmTest(t, inputRead, "Halo Dunia")

	// 3. Read Partial (Stream)
	inputReadPartial := `
    f = buka_file("test_stream.txt", "b");
    p1 = baca(f, 4);
    p2 = baca(f, 1);
    p3 = baca(f, 5);
    tutup_file(f);
    p1 + p3;
    `
	// "Halo" (4) + " " (1) + "Dunia" (5)
	// p1="Halo", p3="Dunia"
	runVmTest(t, inputReadPartial, "HaloDunia")
}
