package discovery

import (
	"path/filepath"
	"testing"

	"github.com/vkrishna03/orchestr/internal/models"
)

func TestSpringBootDetector_Detect(t *testing.T) {
	detector := NewSpringBootDetector()

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "maven project",
			dir:  "testdata/springboot-maven",
			want: true,
		},
		{
			name: "gradle project",
			dir:  "testdata/springboot-gradle",
			want: true,
		},
		{
			name: "vite project",
			dir:  "testdata/vite-npm",
			want: false,
		},
		{
			name: "empty directory",
			dir:  "testdata/empty",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.Detect(tt.dir)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpringBootDetector_Analyze(t *testing.T) {
	detector := NewSpringBootDetector()

	tests := []struct {
		name      string
		dir       string
		wantPort  int
		wantBuild models.BuildTool
	}{
		{
			name:      "maven with yml config",
			dir:       "testdata/springboot-maven",
			wantPort:  8081,
			wantBuild: models.BuildMaven,
		},
		{
			name:      "gradle with properties config",
			dir:       "testdata/springboot-gradle",
			wantPort:  8082,
			wantBuild: models.BuildGradle,
		},
		{
			name:      "maven with properties config",
			dir:       "testdata/springboot-maven-props",
			wantPort:  8085,
			wantBuild: models.BuildMaven,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := detector.Analyze(tt.dir)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}

			if svc.Port != tt.wantPort {
				t.Errorf("Port = %v, want %v", svc.Port, tt.wantPort)
			}

			if svc.BuildTool != tt.wantBuild {
				t.Errorf("BuildTool = %v, want %v", svc.BuildTool, tt.wantBuild)
			}

			if svc.Type != models.ServiceSpringBoot {
				t.Errorf("Type = %v, want %v", svc.Type, models.ServiceSpringBoot)
			}

			if len(svc.StartCmd) == 0 {
				t.Error("StartCmd is empty")
			}

			if svc.HealthURL == "" {
				t.Error("HealthURL is empty")
			}
		})
	}
}

func TestViteDetector_Detect(t *testing.T) {
	detector := NewViteDetector()

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{
			name: "npm project with ts config",
			dir:  "testdata/vite-npm",
			want: true,
		},
		{
			name: "pnpm project with js config",
			dir:  "testdata/vite-pnpm",
			want: true,
		},
		{
			name: "spring boot project",
			dir:  "testdata/springboot-maven",
			want: false,
		},
		{
			name: "empty directory",
			dir:  "testdata/empty",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.Detect(tt.dir)
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestViteDetector_Analyze(t *testing.T) {
	detector := NewViteDetector()

	tests := []struct {
		name      string
		dir       string
		wantPort  int
		wantBuild models.BuildTool
	}{
		{
			name:      "npm project",
			dir:       "testdata/vite-npm",
			wantPort:  3000,
			wantBuild: models.BuildNPM,
		},
		{
			name:      "pnpm project",
			dir:       "testdata/vite-pnpm",
			wantPort:  3001,
			wantBuild: models.BuildPNPM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := detector.Analyze(tt.dir)
			if err != nil {
				t.Fatalf("Analyze() error = %v", err)
			}

			if svc.Port != tt.wantPort {
				t.Errorf("Port = %v, want %v", svc.Port, tt.wantPort)
			}

			if svc.BuildTool != tt.wantBuild {
				t.Errorf("BuildTool = %v, want %v", svc.BuildTool, tt.wantBuild)
			}

			if svc.Type != models.ServiceVite {
				t.Errorf("Type = %v, want %v", svc.Type, models.ServiceVite)
			}

			if len(svc.StartCmd) == 0 {
				t.Error("StartCmd is empty")
			}
		})
	}
}

func TestDiscoverer_DiscoverOne(t *testing.T) {
	discoverer := NewDiscoverer()

	tests := []struct {
		name     string
		dir      string
		wantType models.ServiceType
		wantNil  bool
	}{
		{
			name:     "spring boot maven",
			dir:      "testdata/springboot-maven",
			wantType: models.ServiceSpringBoot,
		},
		{
			name:     "vite npm",
			dir:      "testdata/vite-npm",
			wantType: models.ServiceVite,
		},
		{
			name:    "empty directory",
			dir:     "testdata/empty",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := discoverer.DiscoverOne(tt.dir)
			if err != nil {
				t.Fatalf("DiscoverOne() error = %v", err)
			}

			if tt.wantNil {
				if svc != nil {
					t.Errorf("expected nil, got %+v", svc)
				}
				return
			}

			if svc == nil {
				t.Fatal("expected service, got nil")
			}

			if svc.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", svc.Type, tt.wantType)
			}
		})
	}
}

func TestDiscoverer_Discover(t *testing.T) {
	discoverer := NewDiscoverer()

	t.Run("multi-service directory", func(t *testing.T) {
		services, err := discoverer.Discover("testdata/multi-service")
		if err != nil {
			t.Fatalf("Discover() error = %v", err)
		}

		if len(services) != 2 {
			t.Errorf("expected 2 services, got %d", len(services))
		}

		// Check we found both types
		var foundSpringBoot, foundVite bool
		for _, svc := range services {
			switch svc.Type {
			case models.ServiceSpringBoot:
				foundSpringBoot = true
			case models.ServiceVite:
				foundVite = true
			}
		}

		if !foundSpringBoot {
			t.Error("Spring Boot service not found")
		}
		if !foundVite {
			t.Error("Vite service not found")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		services, err := discoverer.Discover("testdata/empty")
		if err != nil {
			t.Fatalf("Discover() error = %v", err)
		}

		if len(services) != 0 {
			t.Errorf("expected 0 services, got %d", len(services))
		}
	})
}

func TestDiscoverer_SupportedTypes(t *testing.T) {
	discoverer := NewDiscoverer()
	types := discoverer.SupportedTypes()

	if len(types) != 2 {
		t.Errorf("expected 2 types, got %d", len(types))
	}

	expected := map[models.ServiceType]bool{
		models.ServiceSpringBoot: true,
		models.ServiceVite:       true,
	}

	for _, typ := range types {
		if !expected[typ] {
			t.Errorf("unexpected type: %v", typ)
		}
	}
}

func TestSpringBootDetector_PortExtraction(t *testing.T) {
	detector := NewSpringBootDetector()

	tests := []struct {
		name     string
		dir      string
		wantPort int
	}{
		{
			name:     "yml config",
			dir:      "testdata/springboot-maven",
			wantPort: 8081,
		},
		{
			name:     "properties config",
			dir:      "testdata/springboot-gradle",
			wantPort: 8082,
		},
		{
			name:     "env placeholder with default",
			dir:      "testdata/springboot-nested",
			wantPort: 8090,
		},
		{
			name:     "profile-specific config",
			dir:      "testdata/springboot-profile",
			wantPort: 9000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absDir, _ := filepath.Abs(tt.dir)
			port := detector.extractPort(absDir)
			if port != tt.wantPort {
				t.Errorf("extractPort() = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestViteDetector_PortExtraction(t *testing.T) {
	detector := NewViteDetector()

	tests := []struct {
		name     string
		dir      string
		wantPort int
	}{
		{
			name:     "ts config with port",
			dir:      "testdata/vite-npm",
			wantPort: 3000,
		},
		{
			name:     "js config with port",
			dir:      "testdata/vite-pnpm",
			wantPort: 3001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absDir, _ := filepath.Abs(tt.dir)
			port := detector.extractPort(absDir)
			if port != tt.wantPort {
				t.Errorf("extractPort() = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestViteDetector_PackageManagerDetection(t *testing.T) {
	detector := NewViteDetector()

	tests := []struct {
		name     string
		dir      string
		wantTool models.BuildTool
	}{
		{
			name:     "npm lockfile",
			dir:      "testdata/vite-npm",
			wantTool: models.BuildNPM,
		},
		{
			name:     "pnpm lockfile",
			dir:      "testdata/vite-pnpm",
			wantTool: models.BuildPNPM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absDir, _ := filepath.Abs(tt.dir)
			tool := detector.detectPackageManager(absDir)
			if tool != tt.wantTool {
				t.Errorf("detectPackageManager() = %v, want %v", tool, tt.wantTool)
			}
		})
	}
}
