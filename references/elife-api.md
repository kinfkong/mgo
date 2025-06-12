# MGO Usages in elife-api

---
### `/Users/jinggangwang/tc/LINK/elife-api/utils/db.go`

```go
// InitDB initializes the global database session
func InitDB() error {
	url := config.GetConfig().GetString("mongodb.url")
	iotDataUrl := config.GetConfig().GetString("mongodb.iot-data-db-url")

	if url != "" {
		err := initDB(url, &gSession)
		if err != nil {
			log.Errorf("Failed to init user db: %s", err.Error())
			return err
		}
	}
	if iotDataUrl != "" {
		err := initDB(iotDataUrl, &gIoTDataSession)
		if err != nil {
			log.Errorf("Failed to init iot db: %s", err.Error())
			return err
		}
	}
	gStats.initStats()
	return nil
}

func initDB(url string, sessionTobe **mgo.Session) error {
	session, err := mgo.Dial(url)
	if err != nil {
		return err
	}
	session.SetMode(mgo.PrimaryPreferred, true)
	*sessionTobe = session
	return nil
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/temppermission/service.go`

```go
func (service *Service) fetchAutoOccupiedTempPermissions() ([]TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"handledOccupied": true}},
		ReturnNew: false,
	}
	_, err := c.Find(condition).Apply(change, &tempPermission)
	// ... existing code ...
}

func (service *Service) ListTempPermissions(userID string, familyID string, tempPermissionType *string) ([]TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	query := bson.M{
		"userId":   bson.ObjectIdHex(userID),
		"familyId": bson.ObjectIdHex(familyID),
		"type":     typeValue,
	}
	err := c.Find(query).All(&tempPermissions)
	// ... existing code ...
}

func (service *Service) ListAllTempPermissions(userID string, familyID string) ([]TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	query := bson.M{
		"userId":   bson.ObjectIdHex(userID),
		"familyId": bson.ObjectIdHex(familyID),
	}
	err = c.Find(query).All(&tempPermissions)
	// ... existing code ...
}

func (service *Service) CreateTempPermission(userID bson.ObjectId, tempPermission *TempPermission) (*TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	err = c.Insert(tempPermission)
	// ... existing code ...
}

func (service *Service) GetTempPermission(userID string, tempPermissionID string) (*TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id": bson.ObjectIdHex(tempPermissionID),
	}).One(&tempPermission)
	// ... existing code ...
}

func (service *Service) UpdateTempPermission(tempPermission *TempPermission) (*TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    tempPermission.ID,
		"userId": tempPermission.UserID,
	}).One(&existingTempPermission)
	// ... existing code ...
	err = c.Update(bson.M{
		"_id":    tempPermission.ID,
		"userId": tempPermission.UserID,
	}, existingTempPermission)
	// ... existing code ...
}

func (service *Service) validateTempPermission(tempPermission *TempPermission) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	query := bson.M{
		"familyId": tempPermission.FamilyID,
	}
	err := c.Find(query).All(&tempPermissions)
	// ... existing code ...
}

func (service *Service) DeleteTempPermission(userID string, tempPermissionID string) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)

	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    bson.ObjectIdHex(tempPermissionID),
		"userId": bson.ObjectIdHex(userID),
	}).One(&existingTempPermission)
	// ... existing code ...
	err = c.Remove(bson.M{
		"_id":    bson.ObjectIdHex(tempPermissionID),
		"userId": bson.ObjectIdHex(userID),
	})
	// ... existing code ...
}

func (service *Service) SendTempPermission(userID bson.ObjectId, email string, tempPermissionID string, sendType string) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	fc := session.DB("").C(utils.ELifeFamilyCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    bson.ObjectIdHex(tempPermissionID),
		"userId": userID,
	}).One(&existingTempPermission)
	// ... existing code ...
	err = fc.Find(bson.M{
		"_id": existingTempPermission.FamilyID,
	}).One(&familyItem)
	// ... existing code ...
	err = c.Update(bson.M{
		"_id": existingTempPermission.ID,
	}, existingTempPermission)
	// ... existing code ...
}

func (service *Service) Login(tempPermissionID bson.ObjectId, password string) (*TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	err := c.Find(bson.M{"_id": tempPermissionID}).One(&tempPermission)
	// ... existing code ...
	_ = c.Update(bson.M{"_id": tempPermission.ID}, bson.M{"$set": bson.M{
		"loginAt": tempPermission.LoginAt}})
	// ... existing code ...
}

func (service *Service) GetTempPermissionByToken(token string) (*TempPermission, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeTempPermissionCollection)
	// ... existing code ...
	err := c.Find(bson.M{"token": token}).One(&tempPermission)
	// ... existing code ...
}

func (service *Service) checkFamilyOccupied(familyID bson.ObjectId) (bool, error) {
	// ... existing code ...
	fc := session.DB("").C(utils.ELifeFamilyCollection)
	find, err := fc.Find(bson.M{"_id": familyID, "detailSetting": "occupied"}).Count()
	// ... existing code ...
}

func (service *Service) checkFamilyAutolock(familyID bson.ObjectId) (bool, error) {
	// ... existing code ...
	fc := session.DB("").C(utils.ELifeFamilyCollection)
	find, err := fc.Find(bson.M{"_id": familyID, "detailSetting": "autolock"}).Count()
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/temppermission/temppermission.go`

```go
// EnsureIndices create the indexes
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.ELifeTempPermissionCollection)

	index := mgo.Index{
		Key: []string{"userId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"token"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"handledOccupied", "occupiedAt", "disabled"},
	}
	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	return nil
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/schedule/schedule_service.go`

```go
func (service *Service) CreateSchedule(schedule *Schedule) (*Schedule, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	err = c.Insert(schedule)
	// ... existing code ...
}
```

```go
func (service *Service) checkSchedules() error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"taskPickedAt": time.Now()}},
		ReturnNew: false,
	}
	_, err := c.Find(condition).Apply(change, &schedule)

	if err == mgo.ErrNotFound {
		// ... existing code ...
		c.UpdateAll(bson.M{"taskPickedAt": bson.M{
			"$lte": time.Now().Add(time.Duration(-1) * time.Minute)}}, bson.M{
			"$set": bson.M{"taskPickedAt": nil},
			"$inc": bson.M{"failedTimes": 1}})
		return nil
	}
	// ... existing code ...
	if schedule.FailedTimes >= 3 {
		// ... existing code ...
		c.UpdateId(schedule.ID, bson.M{"$set": bson.M{
			"taskPickedAt":  nil,
			"failedTimes":   0,
			"nextTriggerAt": calculateNextScheduleTime(&schedule)}})
		return nil
	}
	// ... existing code ...
}
```

```go
func (service *Service) RemoveAllSchedulesByTag(userID string, familyID bson.ObjectId, tag string) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	_, err = c.RemoveAll(query)
	// ... existing code ...
}
```

```go
func (service *Service) GetSchedulesByTagAndDeviceID(tag string, deviceID bson.ObjectId) ([]Schedule, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	err := c.Find(query).All(&schedules)
	// ... existing code ...
}
```

```go
func (service *Service) ListSchedules(userID string, deviceID *string, panelID *string, familyID *string, tag *string, withHidden bool) ([]Schedule, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	err := c.Find(query).All(&schedules)
	// ... existing code ...
}
```

```go
func (service *Service) GetSchedule(userID string, scheduleID string) (*Schedule, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id": bson.ObjectIdHex(scheduleID),
	}).One(&schedule)
	// ... existing code ...
}
```

```go
func (service *Service) DisableClockTimer(clockTimerID bson.ObjectId) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeClockTimerCollection)
	err := c.Update(bson.M{
		"_id": clockTimerID,
	}, bson.M{
		"$set": bson.M{
			"enabled":   false,
			"updatedAt": time.Now(),
		},
	})
	// ... existing code ...
}
```

```go
func (service *Service) UpdateSchedule(schedule *Schedule) (*Schedule, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    schedule.ID,
		"userId": schedule.UserID,
	}).One(existingSchedule)
	// ... existing code ...
	err = c.Update(bson.M{
		"_id":    schedule.ID,
		"userId": schedule.UserID,
	}, *existingSchedule)
	// ... existing code ...
}
```

```go
func (service *Service) DeleteSchedule(userID string, scheduleID string) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	err := c.Remove(bson.M{
		"_id":    bson.ObjectIdHex(scheduleID),
		"userId": bson.ObjectIdHex(userID),
	})
	// ... existing code ...
}
```

```go
func (service *Service) getAllScenes(schedule *Schedule) ([]scene.BriefScene, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeSceneCollection)
	// ... existing code ...
	err := c.Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"_id": 1, "name": 1, "familyId": 1, "actions": 1}).All(&scenes)
	// ... existing code ...
}
```

```go
func (service *Service) getAllPannels(schedule *Schedule, scenes []scene.BriefScene) ([]panel.BriefPanel, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifePanelCollection)
	// ... existing code ...
	err := c.Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"_id": 1, "name": 1, "familyId": 1, "roomId": 1, "deviceId": 1, "customCodes": 1}).All(&panels)
	// ... existing code ...
}
```

```go
func (service *Service) getAllDevices(schedule *Schedule, scenes []scene.BriefScene, panels []panel.BriefPanel) ([]devicepkg.BriefDevice, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeDeviceCollection)
	// ... existing code ...
	err := c.Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"_id": 1, "name": 1, "familyId": 1, "roomId": 1, "did": 1, "pid": 1, "vendor": 1}).All(&devices)
	// ... existing code ...
}
```

```go
func (service *Service) getAllRooms(schedule *Schedule, scenes []scene.BriefScene, panels []panel.BriefPanel, devices []devicepkg.BriefDevice) ([]room.BriefRoom, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	// ... existing code ...
	err := c.Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"_id": 1, "name": 1, "familyId": 1}).All(&rooms)
	// ... existing code ...
}
```

```go
func (service *Service) getAllFamilies(schedule *Schedule, scenes []scene.BriefScene, panels []panel.BriefPanel, devices []devicepkg.BriefDevice, rooms []room.BriefRoom) ([]family.BriefFamily, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeFamilyCollection)
	// ... existing code ...
	err := c.Find(bson.M{"_id": bson.M{"$in": ids}}).Select(bson.M{"_id": 1, "name": 1}).All(&families)
	// ... existing code ...
}
```

```go
func (service *Service) addScheduleActionLog(schedule *Schedule) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeActionTriggredLogCollection)
	err = c.Insert(actionTriggeredLog)
	// ... existing code ...
}
```

```go
func (service *Service) addScheduleLog(schedueLog *ScheduleLog) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleLogCollection)
	err := c.Insert(schedueLog)
	// ... existing code ...
}
```

```go
func (service *Service) triggerSchedule(schedule *Schedule) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeScheduleCollection)
	// ... existing code ...
	c.UpdateId(schedule.ID, bson.M{
		"$set": bson.M{"taskPickedAt": nil},
		"$inc": bson.M{"failedTimes": 1}})
	// ... existing code ...
	c.UpdateId(schedule.ID, bson.M{"$set": updateFields})
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/schedule/schedule_event_handler.go`

```go
func onPanelDeleted(panel *panel.Panel) error {
	session := utils.NewDBSession()
	defer session.Close()

	// remove the schedule sof the device
	cs := session.DB("").C(utils.ELifeScheduleCollection)
	_, err := cs.RemoveAll(bson.M{
		"actions.panelId": panel.ID,
	})
	if err != nil && err != mgo.ErrNotFound {
		log.Errorf("error delete the panel of panel: %v, error: %v", panel.ID, err)
		return errors.CreateError(500, "delete_schedule_error")
	}

	return nil
}
```

```go
func onDeviceDeleted(device *device.Device) error {
	session := utils.NewDBSession()
	defer session.Close()

	// remove the schedules of the device
	c := session.DB("").C(utils.ELifeScheduleCollection)
	_, err := c.RemoveAll(bson.M{
		"actions.deviceId": device.ID,
	})
	if err != nil && err != mgo.ErrNotFound {
		log.Errorf("error delete the panel of device: %v, error: %v", device.ID, err)
		return errors.CreateError(500, "delete_schedule_error")
	}
	return nil
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/schedule/schedule.go`

```go
// EnsureIndices create the indexes
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.ELifeScheduleCollection)
	slc := session.DB("").C(utils.ELifeScheduleLogCollection)

	index := mgo.Index{
		Key: []string{"nextTriggerAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"taskPickedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"deviceId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"actions.deviceId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"panelId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"actions.panelId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"actions.sceneId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"familyId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"tag"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"scheduleId"},
	}

	if err := slc.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/room/room_service.go`

```go
func (service *Service) ListRooms(userID string, familyID string) ([]Room, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	// ... existing code ...
	err := c.Find(query).All(&rooms)
	// ... existing code ...
}
```

```go
func (service *Service) GetRoom(userID string, roomID string) (*Room, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id": bson.ObjectIdHex(roomID),
	}).One(&room)
	// ... existing code ...
}
```

```go
func (service *Service) CreateRoom(room *Room) (*Room, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	room.ID = bson.NewObjectId()
	err := c.Insert(room)
	// ... existing code ...
}
```

```go
func (service *Service) UpdateRoom(room *Room) (*Room, error) {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    room.ID,
		"userId": room.UserID,
	}).One(&existingRoom)
	// ... existing code ...
	err = c.Update(bson.M{
		"_id":    existingRoom.ID,
		"userId": existingRoom.UserID,
	}, existingRoom)
	// ... existing code ...
}
```

```go
func (service *Service) DeleteRoom(userID string, roomID string) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    bson.ObjectIdHex(roomID),
		"userId": bson.ObjectIdHex(userID),
	}).One(&room)
	// ... existing code ...
	err = c.Remove(bson.M{
		"_id":    bson.ObjectIdHex(roomID),
		"userId": bson.ObjectIdHex(userID),
	})
	// ... existing code ...
}
```

```go
func (service *Service) CheckRoomInFamily(userID bson.ObjectId, roomID bson.ObjectId, familyID bson.ObjectId) error {
	// ... existing code ...
	c := session.DB("").C(utils.ELifeRoomCollection)
	count, err := c.Find(bson.M{
		"_id":      roomID,
		"familyId": familyID,
	}).Count()
	// ... existing code ...
			count, err = c.Find(bson.M{
				"_id":      roomID,
				"familyId": qrlink.SourceFamilyID,
			}).Count()
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/trigger/trigger_service.go`

```go
func (service *Service) FindTriggers(vendor string, pid string, did string, collectionName string, triggers interface{}) error {
	// ... existing code ...
	session := utils.NewDBSession()
	defer session.Close()
	d := session.DB("").C(utils.ELifeDeviceCollection)

	device := device.Device{}
	err = d.Find(bson.M{
		"vendor": vendor,
		"pid":    pid,
		"did":    did,
	}).One(&device)
	// ... existing code ...
	c := session.DB("").C(collectionName)
	query := bson.M{
		"$or": []bson.M{
			{
				"conditions.deviceId": device.ID,
			},
			{
				"secondConditions.deviceId": device.ID,
			},
		},
		"enabled": true,
	}
	err = c.Find(query).All(triggers)
	// ... existing code ...
}
```

```go
func (service *Service) CleanCacheForDevices(collectionName string, deviceIDs []bson.ObjectId) error {
	// ... existing code ...
	d := session.DB("").C(utils.ELifeDeviceCollection)

	devices := make([]device.Device, 0)
	err := d.Find(bson.M{
		"_id": bson.M{
			"$in": deviceIDs,
		},
	}).All(&devices)
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/app/service.go`

```go
func (service *Service) CreateFeedback(feedback *Feedback) (*Feedback, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.ELifeFeedbackCollection)
	if feedback.Rating > 5 || feedback.Rating < 0 {
		return nil, errors.CreateError(400, "invalid_rating")
	}
	feedback.CreatedAt = time.Now()
	feedback.ID = bson.NewObjectId()
	err := c.Insert(feedback)
	if err != nil {
		log.Errorf("failed to create feedback: %v", err)
		return nil, errors.CreateError(500, "internal_error")
	}
	go service.sendEmail(feedback)
	return feedback, nil
}
```

```go
func (service *Service) GetLatestRelease(platform string) (*Release, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.ELifeReleaseCollection)
	var release Release
	err := c.Find(bson.M{
		"platform": platform,
	}).Sort("-createdAt").Limit(1).One(&release)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.CreateError(404, "release_not_found")
		}
		log.Errorf("failed to get latest release: %v", err)
		return nil, errors.CreateError(500, "internal_error")
	}
	return &release, nil
}
```

```go
func (service *Service) CreateRelease(release *Release) (*Release, error) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.ELifeReleaseCollection)
	release.CreatedAt = bson.NewObjectId()
	release.ID = bson.NewObjectId()
	err := c.Insert(release)
	if err != nil {
		log.Errorf("failed to create release: %v", err)
		return nil, errors.CreateError(500, "internal_error")
	}
	return release, nil
}
```

---

### `/Users/jinggangwang/tc/LINK/elife-api/elife/panel/panel_event_handler.go`

```go
func onRoomDeleted(room *room.Room) {
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.ELifePanelCollection)

	_, err := c.UpdateAll(bson.M{
		"roomId": room.ID,
	}, bson.M{
		"$set": bson.M{
			"roomId": nil,
		},
	})
	if err != nil {
		log.Errorf("failed to update room to null. roomID: %s, error: %v",
			room.ID.Hex(), err)
	}
}
```

```go
func onDeviceDeleted(device *device.Device) error {
	session := utils.NewDBSession()
	defer session.Close()

	// remove the panels of the device
	c := session.DB("").C(utils.ELifePanelCollection)
	panels := make([]Panel, 0)

	err := c.Find(bson.M{
		"deviceId": device.ID,
	}).All(&panels)

	if err != nil {
		log.Errorf("error delete the panel of device: %v, error: %v", device.ID, err)
		return errors.CreateError(500, "find_panel_error")
	}

	_, err = c.RemoveAll(bson.M{
		"deviceId": device.ID,
	})

	if err != nil && err != mgo.ErrNotFound {
		log.Errorf("error delete the panel of device: %v, error: %v", device.ID, err)
		return errors.CreateError(500, "delete_panel_error")
	}

	// because the panel is deleted, trigger the panel removed event
	for _, panel := range panels {
		event.GetService().Publish(event.EventPanelDeleted, &panel)
	}
	return nil
}
```

```go
func onDeviceUnLinked(qrLink *qrlink.QrLink, device *device.Device) error {
	session := utils.NewDBSession()
	defer session.Close()

	// remove the panels of the device
	c := session.DB("").C(utils.ELifePanelCollection)
	panels := make([]Panel, 0)

	err := c.Find(bson.M{
		"deviceId": device.ID,
	}).All(&panels)

	if err != nil {
		log.Errorf("error find the panel of device: %v, error: %v", device.ID, err)
		return errors.CreateError(500, "find_panel_error")
	}
	for _, panel := range panels {
		if panel.UserID.Hex() == qrLink.OwnerID.Hex() {
			// the panel is created by owner, trigger an unlink event
			event.GetService().Publish(event.EventPanelUnLinked, qrLink, &panel)
		} else {
			// delete the panel
			err = c.Remove(bson.M{
				"_id": panel.ID,
			})
			if err != nil {
				log.Errorf("error delete the panel of device: %v, error: %v", device.ID, err)
				// ignore the error to continue
			} else {
				// publish the deleted event
				event.GetService().Publish(event.EventPanelDeleted, &panel)
			}
		}
	}
	return nil
}
```
