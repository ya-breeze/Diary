import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { DiaryService } from "./diary.service";
import {
  DiaryItem,
  DiaryItemRequest,
  DiaryItemsListResponse,
} from "../../shared/models";
import { environment } from "../../../environments/environment";
import { ConfigService } from "./config.service";

describe("DiaryService", () => {
  let service: DiaryService;
  let httpMock: HttpTestingController;

  const mockDiaryItem: DiaryItem = {
    date: "2024-01-15",
    title: "Test Entry",
    body: "Test body content",
    tags: ["test", "diary"],
    previousDate: "2024-01-14",
    nextDate: "2024-01-16",
  };

  const mockListResponse: DiaryItemsListResponse = {
    items: [mockDiaryItem],
    totalCount: 1,
  };

  beforeEach(() => {
    const mockConfigService = {
      getApiUrl: () => environment.apiUrl,
      getConfig: () => ({ apiUrl: environment.apiUrl }),
    };

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [{ provide: ConfigService, useValue: mockConfigService }],
    });

    service = TestBed.inject(DiaryService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should get items with date parameter", (done) => {
    const date = "2024-01-15";

    service.getItems(date).subscribe((response) => {
      expect(response).toEqual(mockListResponse);
      done();
    });

    const req = httpMock.expectOne(
      (req) =>
        req.url === `${environment.apiUrl}/items` &&
        req.params.get("date") === date
    );
    expect(req.request.method).toBe("GET");
    req.flush(mockListResponse);
  });

  it("should get items with search parameter", (done) => {
    const search = "test query";

    service.getItems(undefined, search).subscribe((response) => {
      expect(response).toEqual(mockListResponse);
      done();
    });

    const req = httpMock.expectOne(
      (req) =>
        req.url === `${environment.apiUrl}/items` &&
        req.params.get("search") === search
    );
    expect(req.request.method).toBe("GET");
    req.flush(mockListResponse);
  });

  it("should get items with tags parameter", (done) => {
    const tags = ["tag1", "tag2"];

    service.getItems(undefined, undefined, tags).subscribe((response) => {
      expect(response).toEqual(mockListResponse);
      done();
    });

    const req = httpMock.expectOne(
      (req) =>
        req.url === `${environment.apiUrl}/items` &&
        req.params.get("tags") === "tag1,tag2"
    );
    expect(req.request.method).toBe("GET");
    req.flush(mockListResponse);
  });

  it("should get item by date and update current item", (done) => {
    const date = "2024-01-15";

    service.getItemByDate(date).subscribe(() => {
      expect(service.getCurrentItem()).toEqual(mockDiaryItem);
      expect(service.currentDate()).toBe(date);
      done();
    });

    const req = httpMock.expectOne(
      (req) =>
        req.url === `${environment.apiUrl}/items` &&
        req.params.get("date") === date
    );
    req.flush(mockListResponse);
  });

  it("should handle empty item returned by backend for non-existent date", (done) => {
    const date = "2024-01-20";
    const emptyItemResponse: DiaryItemsListResponse = {
      items: [
        {
          date: date,
          title: "",
          body: "",
          tags: [],
          previousDate: "2024-01-14",
          nextDate: "2024-01-16",
        },
      ],
      totalCount: 1,
    };

    service.getItemByDate(date).subscribe(() => {
      const currentItem = service.getCurrentItem();
      expect(currentItem?.date).toBe(date);
      expect(currentItem?.title).toBe("");
      expect(currentItem?.body).toBe("");
      expect(currentItem?.tags).toEqual([]);
      expect(currentItem?.previousDate).toBe("2024-01-14");
      expect(currentItem?.nextDate).toBe("2024-01-16");
      done();
    });

    const req = httpMock.expectOne(
      (req) =>
        req.url === `${environment.apiUrl}/items` &&
        req.params.get("date") === date
    );
    req.flush(emptyItemResponse);
  });

  it("should save item and update current item", (done) => {
    const itemRequest: DiaryItemRequest = {
      date: "2024-01-15",
      title: "New Entry",
      body: "New content",
      tags: ["new"],
    };

    service.saveItem(itemRequest).subscribe((savedItem) => {
      expect(savedItem).toEqual(mockDiaryItem);
      expect(service.getCurrentItem()).toEqual(mockDiaryItem);
      done();
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/items`);
    expect(req.request.method).toBe("PUT");
    expect(req.request.body).toEqual(itemRequest);
    req.flush(mockDiaryItem);
  });

  it("should search items with text and tags", (done) => {
    const searchText = "test";
    const tags = ["diary"];

    service.searchItems(searchText, tags).subscribe((response) => {
      expect(response).toEqual(mockListResponse);
      done();
    });

    const req = httpMock.expectOne(
      (req) =>
        req.url === `${environment.apiUrl}/items` &&
        req.params.get("search") === searchText &&
        req.params.get("tags") === "diary"
    );
    expect(req.request.method).toBe("GET");
    req.flush(mockListResponse);
  });

  it("should return current item", () => {
    service.getItemByDate("2024-01-15").subscribe();

    const req = httpMock.expectOne(
      (req) => req.url === `${environment.apiUrl}/items`
    );
    req.flush(mockListResponse);

    const currentItem = service.getCurrentItem();
    expect(currentItem).toEqual(mockDiaryItem);
  });

  it("should initialize with today's date", () => {
    const today = new Date().toISOString().split("T")[0];
    expect(service.currentDate()).toBe(today);
  });
});
