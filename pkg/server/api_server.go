package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/vobbilis/codegen/server-discovery/pkg/controller"
	"github.com/vobbilis/codegen/server-discovery/pkg/database"
	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

type APIServer struct {
	config        *models.Config
	db            *database.Database
	router        *mux.Router
	discoveryCtrl *controller.DiscoveryController
}

func NewAPIServer(config *models.Config, db *database.Database, discoveryCtrl *controller.DiscoveryController) *APIServer {
	server := &APIServer{
		config:        config,
		db:            db,
		router:        mux.NewRouter(),
		discoveryCtrl: discoveryCtrl,
	}

	server.setupRoutes()
	return server
}

func (s *APIServer) setupRoutes() {
	s.router.HandleFunc("/api/stats", s.handleGetStats).Methods("GET")
	s.router.HandleFunc("/api/servers", s.handleGetServers).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}", s.handleGetServerByID).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/discoveries", s.handleGetServerDiscoveries).Methods("GET")
	s.router.HandleFunc("/api/discoveries", s.handleGetAllDiscoveries).Methods("GET")
	s.router.HandleFunc("/api/discoveries/{id}", s.handleGetDiscoveryByID).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/open-ports", s.handleGetServerOpenPorts).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/ip-addresses", s.handleGetServerIPAddresses).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/installed-software", s.handleGetServerInstalledSoftware).Methods("GET")
	s.router.HandleFunc("/api/servers/{id}/filesystems", s.handleGetServerFilesystems).Methods("GET")
	s.router.HandleFunc("/api/server-tags", s.handleGetServerTags).Methods("GET")
	s.router.HandleFunc("/api/query", s.handleSQLQuery).Methods("POST")

	// Print registered routes for debugging
	s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			methods, _ := route.GetMethods()
			log.Printf("Route: %s [%v]", pathTemplate, methods)
		}
		return nil
	})
}

func (s *APIServer) Start() error {
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{s.config.API.AllowedOrigins},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}).Handler(s.router)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.API.Port),
		Handler:      handler,
		ReadTimeout:  s.config.API.ReadTimeout,
		WriteTimeout: s.config.API.WriteTimeout,
	}

	log.Printf("Starting API server on port %d", s.config.API.Port)
	return srv.ListenAndServe()
}

func (s *APIServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	// Get all servers from database
	servers, err := s.db.GetAllServers()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"server_count":        len(servers),
		"region_distribution": make(map[string]int),
	}

	// Get all discoveries
	discoveries, err := s.db.GetAllDiscoveries()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	stats["discovery_count"] = len(discoveries)

	// Calculate success rate
	successCount := 0
	for _, d := range discoveries {
		if d.Success {
			successCount++
		}
	}
	if len(discoveries) > 0 {
		stats["success_rate"] = float64(successCount) / float64(len(discoveries)) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	// Calculate region distribution
	regionDist := make(map[string]int)
	for _, server := range servers {
		region := server.Region
		if region == "" {
			region = "Unknown"
		}
		regionDist[region]++
	}
	stats["region_distribution"] = regionDist

	respondWithJSON(w, http.StatusOK, stats)
}

func (s *APIServer) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.db.GetAllServers()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, servers)
}

func (s *APIServer) handleGetServerDiscoveries(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid server ID"})
		return
	}

	discoveries, err := s.db.GetServerDiscoveries(strconv.Itoa(serverID))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, discoveries)
}

func (s *APIServer) handleGetServerTags(w http.ResponseWriter, r *http.Request) {
	// Get all unique tags from the database
	tags, err := s.db.GetAllServerTags()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Transform tags into a map of tag names to tag values
	uniqueTags := make(map[string][]string)
	for _, tag := range tags {
		uniqueTags[tag.TagName] = append(uniqueTags[tag.TagName], tag.TagValue)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uniqueTags)
}

func (s *APIServer) handleGetServerByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid server ID"})
		return
	}

	server, err := s.db.GetServerDetails(strconv.Itoa(serverID))
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
			return
		}
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, server)
}

func (s *APIServer) handleGetServerOpenPorts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid server ID"})
		return
	}

	ports, err := s.db.GetServerOpenPorts(strconv.Itoa(serverID))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, ports)
}

func (s *APIServer) handleGetServerIPAddresses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid server ID"})
		return
	}

	ipAddresses, err := s.db.GetServerIPAddresses(strconv.Itoa(serverID))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, ipAddresses)
}

func (s *APIServer) handleGetServerInstalledSoftware(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid server ID"})
		return
	}

	software, err := s.db.GetServerInstalledSoftware(strconv.Itoa(serverID))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, software)
}

func (s *APIServer) handleGetServerFilesystems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid server ID"})
		return
	}

	filesystems, err := s.db.GetServerFilesystems(strconv.Itoa(serverID))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, filesystems)
}

func (s *APIServer) handleGetAllDiscoveries(w http.ResponseWriter, r *http.Request) {
	discoveries, err := s.db.GetAllDiscoveries()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, discoveries)
}

func (s *APIServer) handleGetDiscoveryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	discoveryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid discovery ID"})
		return
	}

	discovery, err := s.db.GetDiscoveryByID(discoveryID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w, http.StatusNotFound, map[string]string{"error": "Discovery not found"})
			return
		}
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondWithJSON(w, http.StatusOK, discovery)
}

func (s *APIServer) handleSQLQuery(w http.ResponseWriter, r *http.Request) {
	var query struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if query.Query == "" {
		respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Query is required"})
		return
	}

	// Execute the query
	rows, err := s.db.Query(query.Query)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Prepare slice for values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan results
	var results []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	respondWithJSON(w, http.StatusOK, results)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
