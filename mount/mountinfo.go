package mount

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Info per mount
type Info struct {
	MountID        string
	ParentID       string
	MajorMinor     string
	Root           string
	MountPoint     string
	MountOptions   string
	OptionalFields []string
	FilesystemType string
	MountSource    string
	SuperOptions   string
}

// Infos provides current mountinfo mounts within this pids mount namespace
func Infos() ([]Info, error) {
	fh, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	mountInfos := []Info{}

	rdr := bufio.NewReader(fh)

	var isEOF bool
	for {
		line, err := rdr.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		} else if err == io.EOF {
			isEOF = true
		}

		if line != "" {
			m, err := parseMountInfoLine(line)
			if err != nil {
				return nil, err
			}
			mountInfos = append(mountInfos, m)
		}

		if isEOF {
			break
		}
	}

	return mountInfos, nil
}

/*
	From linux/Documentation/proc.txt 3.5

	36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
	(1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

	(1) mount ID:  unique identifier of the mount (may be reused after umount)
	(2) parent ID:  ID of parent (or of self for the top of the mount tree)
	(3) major:minor:  value of st_dev for files on filesystem
	(4) root:  root of the mount within the filesystem
	(5) mount point:  mount point relative to the process's root
	(6) mount options:  per mount options
	(7) optional fields:  zero or more fields of the form "tag[:value]"
	(8) separator:  marks the end of the optional fields
	(9) filesystem type:  name of filesystem of the form "type[.subtype]"
	(10) mount source:  filesystem specific information or "none"
	(11) super options:  per super block options

	Parsers should ignore all unrecognised optional fields.  Currently the                                                                                                                                             possible optional fields are:

	shared:X  mount is shared in peer group X
	master:X  mount is slave to peer group X
	propagate_from:X  mount is slave and receives propagation from peer group X (*)
	unbindable  mount is unbindable
*/
func parseMountInfoLine(line string) (Info, error) {
	info := Info{}
	fields := strings.Fields(line)
	if len(fields) < 10 {
		// as field 7 is optional
		return info, fmt.Errorf("expected at least 10 fields, only got %d", len(fields))
	}
	info.MountID = fields[0]
	info.ParentID = fields[1]
	info.MajorMinor = fields[2]
	info.Root = fields[3]
	info.MountPoint = fields[4]
	info.MountOptions = fields[5]

	info.SuperOptions = fields[len(fields)-1]
	info.MountSource = fields[len(fields)-2]
	info.FilesystemType = fields[len(fields)-3]

	if fields[6] != "-" {
		info.OptionalFields = fields[6 : len(fields)-5]
	}

	if os.Getenv("DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: mountinfo: %#v\n", info)
	}
	return info, nil
}
