package auth

// FacultyPrograms contains a faculty and its study programs.
type FacultyPrograms struct {
	Faculty  string   `json:"faculty"`
	Programs []string `json:"programs"`
}

// Static list used by frontend to render dropdowns for faculty/study program.
var facultyProgramOptions = []FacultyPrograms{
	{
		Faculty: "Fakultas Agama / Syariah",
		Programs: []string{
			"S1 Hukum Keluarga Islam",
		},
	},
	{
		Faculty: "Fakultas Ekonomi",
		Programs: []string{
			"S1 Akuntansi",
			"S1 Manajemen",
		},
	},
	{
		Faculty: "Fakultas Keguruan & Ilmu Pendidikan (FKIP)",
		Programs: []string{
			"S1 Pendidikan Kimia",
			"S1 Pendidikan Bahasa Inggris",
			"S1 Pendidikan Matematika",
			"S1 PG PAUD",
		},
	},
	{
		Faculty: "Fakultas Kesehatan",
		Programs: []string{
			"D3 Kebidanan",
			"D3 Keperawatan",
		},
	},
	{
		Faculty: "Fakultas Teknik",
		Programs: []string{
			"S1 Teknik Informatika",
			"S1 Teknik Industri",
			"S1 Teknik Mesin",
			"S1 Teknik Sipil",
		},
	},
	{
		Faculty: "Fakultas Pertanian",
		Programs: []string{
			"S1 Agroteknologi",
			"S1 Agribisnis",
		},
	},
}
