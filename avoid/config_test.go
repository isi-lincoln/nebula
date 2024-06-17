package avoid

/*
func TestLoadingConfig(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	base := filepath.Dir(filename)
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/configs", base))
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range files {
		test_file := ""
		if strings.Contains(fi.Name(), "yaml") {
			if !strings.Contains(fi.Name(), "swp") {
				test_file = fi.Name()
			} else {
				continue
			}
		} else {
			continue
		}

		configPath := fmt.Sprintf("%s/configs/%s", base, test_file)

		t.Logf("%s\n", configPath)

		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("%v", err)
		} else {
			t.Logf("%#v\n", cfg)
			if cfg.Avoid == nil {
				t.Fatalf("Avoid configuration missing\n")
			}
		}
	}
}
*/
