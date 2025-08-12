package main

import (
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "os"
        "path/filepath"
        "regexp"
        "strings"
        "time"

        "github.com/joho/godotenv"
        "github.com/netisu/aeno"
)

const (
        scale      = 1
        fovy       = 50
        near       = 0.1
        far        = 1000
        amb        = "AAAAAA" // d4d4d4
        lightcolor = "777777" // 696969
        Dimentions = 512      // april fools (15)
)

var (
        eye          = aeno.V(0.75, 0.85, 2)
        center       = aeno.V(0, 0.06, 0)
        up           = aeno.V(0, 1, 0)
        light        = aeno.V(0, 6, 4).Normalize()
        cdnDirectory = "/var/www/cdn" // set this to your storage root
        envDir       = "/var/www/renderer"
)

// hatKeyPattern is a regular expression to match keys like "hat_1", "hat_123", etc.
var hatKeyPattern = regexp.MustCompile(`^hat_\d+$`)

func env(key string) string {
        // Attempt to load .env file from the envDir
        err := godotenv.Load(filepath.Join(envDir, ".env"))
        if err != nil {
                // Fallback to trying to load from the cdnDirectory
                err = godotenv.Load(filepath.Join(cdnDirectory, ".env"))
                if err != nil {
                        log.Println("Warning: .env file not found in executable path or rootDir. Relying on environment variables.")
                }
        }
        return os.Getenv(key)
}

func main() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                if env("POST_KEY") != "" && r.Header.Get("Aeo-Access-Key") != env("POST_KEY") {
                        fmt.Println("Unauthorized request")
                        http.Error(w, "Unauthorized request", http.StatusForbidden)
                        return
                }
                if r.Method != http.MethodGet && r.Method != http.MethodPost {
                        fmt.Println("Method not allowed")
                        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                        return
                }
                renderCommand(w, r)
        })

        // Start the HTTP server
        fmt.Printf("Starting server on %s\n", env("SERVER_ADDRESS"))
        if err := http.ListenAndServe(env("SERVER_ADDRESS"), nil); err != nil {
                fmt.Println("HTTP server error:", err)
        }
}

func renderCommand(w http.ResponseWriter, r *http.Request) {
        renderType := r.URL.Query().Get("RenderType")
        if renderType == "" {
                renderType = r.URL.Query().Get("renderType") // fallback
        }
        if renderType == "" {
                // For backward compatibility, check body if query param is missing
                body, err := io.ReadAll(r.Body)
                if err == nil && len(body) > 0 {
                        var data map[string]interface{}
                        if json.Unmarshal(body, &data) == nil {
                                if rt, ok := data["RenderType"].(string); ok {
                                        renderType = rt
                                }
                        }
                        // Important: Close the body after reading to prevent resource leaks
                        r.Body.Close()
                }
        }

        fmt.Println("Running Function", renderType)
        switch renderType {
        case "user":
                renderUser(w, r)
        case "item":
                renderItemPreview(w, r)
        default:
                fmt.Println("Invalid or missing RenderType:", renderType)
                http.Error(w, "Invalid or missing RenderType", http.StatusBadRequest)
        }
}

func renderUser(w http.ResponseWriter, r *http.Request) {
        hash := r.URL.Query().Get("hash")
        if hash == "" || hash == "default" {
                fmt.Println("Avatar Hash is required for renderUser")
                http.Error(w, "Avatar Hash is required for renderUser", http.StatusBadRequest)
                return
        }

        // Delegate user avatar rendering logic here
        fmt.Println("Getting userstring", hash)

        // Get colors
        head_color := r.URL.Query().Get("head_color")
        if head_color == "" {
                head_color = "f3b700"
        }
        torso_color := r.URL.Query().Get("torso_color")
        if torso_color == "" {
                torso_color = "f3b700"
        }
        leftLeg_color := r.URL.Query().Get("leftLeg_color")
        if leftLeg_color == "" {
                leftLeg_color = "f3b700"
        }
        rightLeg_color := r.URL.Query().Get("rightLeg_color")
        if rightLeg_color == "" {
                rightLeg_color = "f3b700"
        }
        leftArm_color := r.URL.Query().Get("leftArm_color")
        if leftArm_color == "" {
                leftArm_color = "f3b700"
        }
        rightArm_color := r.URL.Query().Get("rightArm_color")
        if rightArm_color == "" {
                rightArm_color = "f3b700"
        }

        // Get items
        hat1 := r.URL.Query().Get("hat_1")
        if hat1 == "" {
                hat1 = "none"
        }
        hat2 := r.URL.Query().Get("hat_2")
        if hat2 == "" {
                hat2 = "none"
        }
        hat3 := r.URL.Query().Get("hat_3")
        if hat3 == "" {
                hat3 = "none"
        }
        hat4 := r.URL.Query().Get("hat_4")
        if hat4 == "" {
                hat4 = "none"
        }
        hat5 := r.URL.Query().Get("hat_5")
        if hat5 == "" {
                hat5 = "none"
        }
        hat6 := r.URL.Query().Get("hat_6")
        if hat6 == "" {
                hat6 = "none"
        }
        face := r.URL.Query().Get("face")
        if face == "" {
                face = "none"
        }
        tool := r.URL.Query().Get("tool")
        if tool == "" {
                tool = "none"
        }
        shirt := r.URL.Query().Get("shirt")
        if shirt == "" {
                shirt = "none"
        }
        tshirt := r.URL.Query().Get("tshirt")
        if tshirt == "" {
                tshirt = "none"
        }
        pants := r.URL.Query().Get("pants")
        if pants == "" {
                pants = "none"
        }

        start := time.Now()
        fmt.Println("Drawing Objects...")
        objects := generateObjects(
                torso_color, leftLeg_color, rightLeg_color, rightArm_color, head_color, face,
                shirt, pants, tshirt,
                hat1, hat2, hat3, hat4, hat5, hat6,
                tool, leftArm_color,
        )

        fmt.Println("Exporting to", cdnDirectory, "thumbnails")
        destPath := filepath.Join(cdnDirectory, "thumbnails", hash) // The hash is the full filename from PHP
        destDir := filepath.Dir(destPath)
        if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
                log.Printf("Failed to create directory %s: %v", destDir, err)
                http.Error(w, "Internal server error", http.StatusInternalServerError)
                return
        }

        aeno.GenerateScene(
                true,
                destPath,
                objects,
                eye,
                center,
                up,
                fovy,
                Dimentions,
                scale,
                light,
                amb,
                lightcolor,
                near,
                far,
        )

        fmt.Println("User render completed in", time.Since(start))
        w.Header().Set("Content-Type", "image/png")
}

func renderItemPreview(w http.ResponseWriter, r *http.Request) {
        itemID := r.URL.Query().Get("item")
        itemHash := r.URL.Query().Get("itemhash")
        itemType := r.URL.Query().Get("itemtype")

        if itemHash == "" || itemHash == "none" || itemID == "" {
                http.Error(w, "Item hash and ID are required for preview", http.StatusBadRequest)
                return
        }
        if itemType == "" {
                http.Error(w, "Item type is required for preview", http.StatusBadRequest)
                return
        }

        // Default avatar values
        torso_color, leftLeg_color, rightLeg_color, rightArm_color, head_color, face,
                shirt, pants, tshirt,
                hat1, hat2, hat3, hat4, hat5, hat6,
                tool, leftArm_color := "f3b700", "f3b700", "f3b700", "f3b700", "f3b700", "none", "none", "none", "none", "none", "none", "none", "none", "none", "none", "none", "f3b700"

        // Apply the item to the correct slot
        switch itemType {
        case "face":
                face = itemID
        case "hat":
                hat1 = itemID
        case "tool":
                tool = itemID
        case "shirt":
                shirt = itemID
        case "tshirt":
                tshirt = itemID
        case "pants":
                pants = itemID
        }

        start := time.Now()
        fmt.Println("Drawing item preview for:", itemHash)
        objects := generateObjects(
                torso_color, leftLeg_color, rightLeg_color, rightArm_color, head_color, face,
                shirt, pants, tshirt,
                hat1, hat2, hat3, hat4, hat5, hat6,
                tool, leftArm_color,
        )

        destPath := filepath.Join(cdnDirectory, "thumbnails", itemHash) // The hash is the full filename
        destDir := filepath.Dir(destPath)
        if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
                log.Printf("Failed to create directory %s: %v", destDir, err)
                http.Error(w, "Internal server error", http.StatusInternalServerError)
                return
        }

        aeno.GenerateScene(
                true,
                destPath,
                objects,
                eye,
                center,
                up,
                fovy,
                Dimentions,
                scale,
                light,
                amb,
                lightcolor,
                near,
                far,
        )

        fmt.Println("Item preview completed in", time.Since(start))
        w.Header().Set("Content-Type", "image/png")
}

func generateObjects(
        torsoColor, leftLegColor, rightLegColor, rightArmColor, headColor, faceID,
        shirtID, pantsID, tshirtID,
        hat1, hat2, hat3, hat4, hat5, hat6,
        toolID, leftArmColor string) []*aeno.Object {

        var allObjects []*aeno.Object
        cdnURL := env("CDN_URL")

        bodyAndApparelObjects := Texturize(torsoColor, rightArmColor, leftLegColor, rightLegColor, shirtID, pantsID, tshirtID)
        allObjects = append(allObjects, bodyAndApparelObjects...)

        // Add Head
        allObjects = append(allObjects, &aeno.Object{
                Mesh:    aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/Head.obj", cdnURL)),
                Color:   aeno.HexColor(headColor),
                Texture: AddFace(faceID),
                Matrix:  aeno.Identity(),
        })

        // Add Hats
        hats := []string{hat1, hat2, hat3, hat4, hat5, hat6}
        for _, hatID := range hats {
                if hatID != "none" && hatID != "" {
                        if obj := RenderItem(hatID); obj != nil {
                                allObjects = append(allObjects, obj)
                        }
                }
        }

        // Add Tool and Left Arm
        armObjects := ToolClause(toolID, leftArmColor, shirtID)
        allObjects = append(allObjects, armObjects...)

        return allObjects
}

// getTextureFromAPI is a helper function to fetch a texture ID from the API and return an aeno.Texture
func getTextureFromAPI(itemID string) aeno.Texture {
        apiUrl := env("API_URL")
        if itemID == "none" || itemID == "" {
                return nil
        }

        // Fetch poly data from the API
        resp, err := http.Get(fmt.Sprintf("%s/v1/assets/getPoly/1/%s", apiUrl, itemID))
        if err != nil {
                log.Printf("Failed to fetch poly data for item %s: %v", itemID, err)
                return nil
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                log.Printf("API returned non-200 status for item %s: %s", itemID, resp.Status)
                return nil
        }

        var data []map[string]string
        err = json.NewDecoder(resp.Body).Decode(&data)
        if err != nil {
                log.Printf("Failed to decode JSON for item %s: %v", itemID, err)
                return nil
        }

        if len(data) == 0 {
                log.Printf("No data returned for item %s", itemID)
                return nil
        }

        textureID := data[0]["texture"]
        textureID = strings.TrimPrefix(textureID, "asset://")

        if textureID == "" {
                log.Printf("Texture ID is empty for item %s.", itemID)
                return nil
        }

        textureURL := fmt.Sprintf("%s/v1/assets/get/%s", apiUrl, textureID)
        fmt.Printf("Loading texture for item %s from URL: %s\n", itemID, textureURL)
        return aeno.LoadTextureFromURL(textureURL)
}

func Texturize(torsoColor, rightArmColor, leftLegColor, rightLegColor, shirtID, pantsID, tshirtID string) []*aeno.Object {
        objects := []*aeno.Object{}
        cdnUrl := env("CDN_URL")

        shirtTexture := getTextureFromAPI(shirtID)
        pantsTexture := getTextureFromAPI(pantsID)
        tshirtTexture := getTextureFromAPI(tshirtID)

        // Torso
        objects = append(objects, &aeno.Object{
                Mesh:    aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/Torso.obj", cdnUrl)),
                Color:   aeno.HexColor(torsoColor),
                Texture: shirtTexture,
                Matrix:  aeno.Identity(),
        })

        // Left Arm
        objects = append(objects, &aeno.Object{
                Mesh:    aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/LeftArm.obj", cdnUrl)),
                Color:   aeno.HexColor(rightArmColor),
                Texture: shirtTexture,
                Matrix:  aeno.Identity(),
        })

        // Legs
        objects = append(objects,
                &aeno.Object{
                        Mesh:    aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/LeftLeg.obj", cdnUrl)),
                        Color:   aeno.HexColor(leftLegColor),
                        Texture: pantsTexture,
                        Matrix:  aeno.Identity(),
                },
                &aeno.Object{
                        Mesh:    aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/RightLeg.obj", cdnUrl)),
                        Color:   aeno.HexColor(rightLegColor),
                        Texture: pantsTexture,
                        Matrix:  aeno.Identity(),
                },
        )

        // T-Shirt Overlay
        if tshirtID != "none" && tshirtID != "" {
                objects = append(objects, &aeno.Object{
                        Mesh:    aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/tshirt.obj", cdnUrl)),
                        Color:   aeno.Transparent,
                        Texture: tshirtTexture,
                        Matrix:  aeno.Identity(),
                })
        }

        return objects
}

func ToolClause(toolID, leftArmColor, shirtID string) []*aeno.Object {
        objects := []*aeno.Object{}
        cdnUrl := env("CDN_URL")

        shirtTexture := getTextureFromAPI(shirtID)

        var armMesh *aeno.Mesh
        if toolID != "none" && toolID != "" {
                armMesh = aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/ArmHold.obj", cdnUrl))
                if toolObj := RenderItem(toolID); toolObj != nil {
                        objects = append(objects, toolObj)
                }
        } else {
                armMesh = aeno.LoadObjectFromURL(fmt.Sprintf("%s/assets/RightArm.obj", cdnUrl))
        }

        armObject := &aeno.Object{
                Mesh:    armMesh,
                Color:   aeno.HexColor(leftArmColor),
                Texture: shirtTexture,
                Matrix:  aeno.Identity(),
        }
        objects = append(objects, armObject)

        return objects
}

func RenderItem(itemID string) *aeno.Object {
        apiUrl := env("API_URL")
        if itemID == "none" || itemID == "" {
                return nil
        }

        // Fetch poly data from the API
        resp, err := http.Get(fmt.Sprintf("%s/v1/assets/getPoly/1/%s", apiUrl, itemID))
        if err != nil {
                log.Printf("Failed to fetch poly data for item %s: %v", itemID, err)
                return nil
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                log.Printf("API returned non-200 status for item %s: %s", itemID, resp.Status)
                return nil
        }

        var data []map[string]string
        err = json.NewDecoder(resp.Body).Decode(&data)
        if err != nil {
                log.Printf("Failed to decode JSON for item %s: %v", itemID, err)
                return nil
        }

        if len(data) == 0 {
                log.Printf("No data returned for item %s", itemID)
                return nil
        }

        meshID := data[0]["mesh"]
        meshID = strings.TrimPrefix(meshID, "asset://")

        textureID := data[0]["texture"]
        textureID = strings.TrimPrefix(textureID, "asset://")

        if meshID == "" {
                log.Printf("Mesh ID is empty for item %s", itemID)
                return nil // Cannot render without a mesh
        }

        meshURL := fmt.Sprintf("%s/v1/assets/get/%s", apiUrl, meshID)
        var texture aeno.Texture
        if textureID != "" {
                textureURL := fmt.Sprintf("%s/v1/assets/get/%s", apiUrl, textureID)
                texture = aeno.LoadTextureFromURL(textureURL)
        }

        return &aeno.Object{
                Mesh:    aeno.LoadObjectFromURL(meshURL),
                Color:   aeno.Transparent,
                Texture: texture,
                Matrix:  aeno.Identity(),
        }
}

func AddFace(faceID string) aeno.Texture {
        apiUrl := env("API_URL")
        cdnURL := env("CDN_URL")

        if faceID == "" || faceID == "none" {
                defaultURL := fmt.Sprintf("%s/assets/DefaultFace.png", cdnURL)
                fmt.Printf("AddFace: No face provided. Using default face texture: %s\n", defaultURL)
                return aeno.LoadTextureFromURL(defaultURL)
        }

        // Fetch poly data from the API to get the texture asset ID
        resp, err := http.Get(fmt.Sprintf("%s/v1/assets/getPoly/1/%s", apiUrl, faceID))
        if err != nil {
                log.Printf("Failed to fetch poly data for face %s: %v", faceID, err)
                return nil
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                log.Printf("API returned non-200 status for face %s: %s", faceID, resp.Status)
                return nil
        }

        var data []map[string]string
        err = json.NewDecoder(resp.Body).Decode(&data)
        if err != nil {
                log.Printf("Failed to decode JSON for face %s: %v", faceID, err)
                return nil
        }

        if len(data) == 0 {
                log.Printf("No data returned for face %s", faceID)
                return nil
        }

        textureID := data[0]["texture"]
        textureID = strings.TrimPrefix(textureID, "asset://")

        if textureID == "" {
                log.Printf("Texture ID is empty for face %s, using default.", faceID)
                defaultURL := fmt.Sprintf("%s/assets/DefaultFace.png", cdnURL)
                return aeno.LoadTextureFromURL(defaultURL)
        }

        // Only PNGs are supported for faces, so we construct the URL directly.
        textureURL := fmt.Sprintf("%s/v1/assets/get/%s", apiUrl, textureID)
        fmt.Printf("AddFace: Loading face texture from URL: %s\n", textureURL)
        return aeno.LoadTextureFromURL(textureURL)
}
