package services

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"transfigurr/constants"
	"transfigurr/interfaces"
	"transfigurr/models"
	"transfigurr/tasks"
)

type ScanService struct {
	scanQueue       []models.Item
	scanSet         map[string]struct{}
	mu              sync.Mutex
	cond            *sync.Cond
	metadataService interfaces.MetadataServiceInterface
	encodeService   interfaces.EncodeServiceInterface
	eventService    interfaces.EventServiceInterface
	seriesRepo      interfaces.SeriesRepositoryInterface
	seasonRepo      interfaces.SeasonRepositoryInterface
	episodeRepo     interfaces.EpisodeRepositoryInterface
	movieRepo       interfaces.MovieRepositoryInterface
	settingRepo     interfaces.SettingRepositoryInterface
	systemRepo      interfaces.SystemRepositoryInterface
	profileRepo     interfaces.ProfileRepositoryInterface
	authRepo        interfaces.AuthRepositoryInterface
	userRepo        interfaces.UserRepositoryInterface
	historyRepo     interfaces.HistoryRepositoryInterface
	eventRepo       interfaces.EventRepositoryInterface
	codecRepo       interfaces.CodecRepositoryInterface
}

func NewScanService(eventService interfaces.EventServiceInterface, metadataService interfaces.MetadataServiceInterface, encodeService interfaces.EncodeServiceInterface, seriesRepo interfaces.SeriesRepositoryInterface, seasonRepo interfaces.SeasonRepositoryInterface, episodeRepo interfaces.EpisodeRepositoryInterface, movieRepo interfaces.MovieRepositoryInterface, settingRepo interfaces.SettingRepositoryInterface, systemRepo interfaces.SystemRepositoryInterface, profileRepo interfaces.ProfileRepositoryInterface, authRepo interfaces.AuthRepositoryInterface, userRepo interfaces.UserRepositoryInterface, historyRepo interfaces.HistoryRepositoryInterface, eventRepo interfaces.EventRepositoryInterface, codecRepo interfaces.CodecRepositoryInterface) interfaces.ScanServiceInterface {
	service := &ScanService{
		scanQueue:       make([]models.Item, 0),
		scanSet:         make(map[string]struct{}),
		metadataService: metadataService,
		encodeService:   encodeService,
		eventService:    eventService,
		seriesRepo:      seriesRepo,
		seasonRepo:      seasonRepo,
		episodeRepo:     episodeRepo,
		movieRepo:       movieRepo,
		settingRepo:     settingRepo,
		systemRepo:      systemRepo,
		profileRepo:     profileRepo,
		authRepo:        authRepo,
		userRepo:        userRepo,
		historyRepo:     historyRepo,
		eventRepo:       eventRepo,
		codecRepo:       codecRepo,
	}
	service.cond = sync.NewCond(&service.mu)
	return service
}

func (s *ScanService) EnqueueAll() {
	s.EnqueueAllMovies()
	s.EnqueueAllSeries()
}

func (s *ScanService) EnqueueAllMovies() {
	movies, err := s.movieRepo.GetMovies()
	if err != nil {
		log.Print(err)
		return
	}
	movieFiles, err := ioutil.ReadDir(constants.MoviesPath)
	if err != nil {
		return
	}
	for _, file := range movieFiles {
		s.Enqueue(models.Item{Id: file.Name(), Type: "movie"})
	}
	for _, movieItem := range movies {
		s.Enqueue(models.Item{Id: movieItem.Id, Type: "movie"})
	}
}

func (s *ScanService) EnqueueAllSeries() {
	series, err := s.seriesRepo.GetSeries()
	if err != nil {
		log.Print(err)
		return
	}
	seriesFiles, err := ioutil.ReadDir(constants.SeriesPath)
	if err != nil {
		return
	}
	for _, file := range seriesFiles {
		s.Enqueue(models.Item{Id: file.Name(), Type: "series"})
	}

	for _, seriesItem := range series {
		s.Enqueue(models.Item{Id: seriesItem.Id, Type: "series"})
	}
}

func (s *ScanService) Enqueue(item models.Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	itemID := fmt.Sprintf("%s_%s", item.Type, item.Id)
	if _, ok := s.scanSet[itemID]; !ok {
		s.scanSet[itemID] = struct{}{}
		s.scanQueue = append(s.scanQueue, item)
		s.cond.Signal()
	}
}

func (s *ScanService) process() {
	for {
		s.mu.Lock()
		for len(s.scanQueue) == 0 {
			s.cond.Wait()
		}
		item := s.scanQueue[0]
		s.scanQueue = s.scanQueue[1:]
		s.mu.Unlock()
		s.processItem(item)
		tasks.ScanSystem(s.seriesRepo, s.systemRepo)
	}
}

func (s *ScanService) processItem(item models.Item) {
	if item.Type == "movie" {
		tasks.ScanMovie(item.Id, s.movieRepo, s.settingRepo, s.profileRepo)
		log.Print("bouta validate", item.Id)
		tasks.ValidateMovie(item.Id, s.movieRepo)
		movie, err := s.movieRepo.GetMovieById(item.Id)
		if err != nil {
			log.Print(err)
			return
		}

		if movie.Name == "" {
			s.eventService.Log("INFO", "scan", "Scanning movie: "+item.Id)
			s.metadataService.Enqueue(models.Item{Type: "movie", Id: movie.Id})
		}
		if movie.Missing && movie.Monitored {
			s.encodeService.Enqueue(models.Item{Type: "movie", Id: movie.Id})
		}
	} else if item.Type == "series" {
		s.eventService.Log("INFO", "scan", "Scanning series: "+item.Id)
		log.Print("Scanning series: " + item.Id)
		tasks.ScanSeries(s.encodeService, item.Id, s.seriesRepo, s.seasonRepo, s.episodeRepo, s.settingRepo, s.profileRepo)
		tasks.ValidateSeries(item.Id, s.seriesRepo, s.seasonRepo, s.episodeRepo)
		series, err := s.seriesRepo.GetSeriesByID(item.Id)
		if err != nil {
			log.Print(err)
		}

		if series.Name == "" {
			s.metadataService.Enqueue(models.Item{Type: "series", Id: series.Id})
		} else {
			for _, season := range series.Seasons {
				for _, episode := range season.Episodes {
					if episode.EpisodeName == "" {
						s.metadataService.Enqueue(models.Item{Type: "series", Id: series.Id})
						break
					}
				}
			}
		}
	}

	s.mu.Lock()
	delete(s.scanSet, fmt.Sprintf("%s_%s", item.Type, item.Id))
	s.mu.Unlock()
}

func (s *ScanService) Startup() {
	go s.process()
}
