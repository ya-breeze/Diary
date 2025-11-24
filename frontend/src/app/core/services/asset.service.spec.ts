import { TestBed } from "@angular/core/testing";
import {
  HttpClientTestingModule,
  HttpTestingController,
} from "@angular/common/http/testing";
import { AssetService } from "./asset.service";
import { AssetsBatchResponse } from "../../shared/models";
import { environment } from "../../../environments/environment";
import { ConfigService } from "./config.service";

describe("AssetService", () => {
  let service: AssetService;
  let httpMock: HttpTestingController;
  let configService: ConfigService;

  beforeEach(() => {
    const mockConfigService = {
      getApiUrl: () => environment.apiUrl,
      getConfig: () => ({ apiUrl: environment.apiUrl }),
    };

    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [{ provide: ConfigService, useValue: mockConfigService }],
    });

    service = TestBed.inject(AssetService);
    httpMock = TestBed.inject(HttpTestingController);
    configService = TestBed.inject(ConfigService);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("should upload a single asset", (done) => {
    const mockFile = new File(["test content"], "test.jpg", {
      type: "image/jpeg",
    });
    const mockPath = "uploads/test-uuid.jpg";

    service.uploadAsset(mockFile).subscribe((path) => {
      expect(path).toBe(mockPath);
      done();
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/assets`);
    expect(req.request.method).toBe("POST");
    expect(req.request.body instanceof FormData).toBe(true);

    const formData = req.request.body as FormData;
    expect(formData.get("asset")).toBe(mockFile);

    req.flush(mockPath);
  });

  it("should upload multiple assets in batch", (done) => {
    const mockFiles = [
      new File(["content1"], "file1.jpg", { type: "image/jpeg" }),
      new File(["content2"], "file2.png", { type: "image/png" }),
    ];

    const mockResponse: AssetsBatchResponse = {
      files: [
        {
          originalName: "file1.jpg",
          savedName: "uploads/file1-uuid.jpg",
          size: 1024,
          contentType: "image/jpeg",
        },
        {
          originalName: "file2.png",
          savedName: "uploads/file2-uuid.png",
          size: 2048,
          contentType: "image/png",
        },
      ],
      count: 2,
    };

    service.uploadAssetsBatch(mockFiles).subscribe((response) => {
      expect(response).toEqual(mockResponse);
      expect(response.files.length).toBe(2);
      expect(response.count).toBe(2);
      done();
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/assets/batch`);
    expect(req.request.method).toBe("POST");
    expect(req.request.body instanceof FormData).toBe(true);

    req.flush(mockResponse);
  });

  it("should generate correct asset URL for web display", () => {
    const path = "test-image.jpg";
    const baseUrl = environment.apiUrl.replace("/v1", "");
    const expectedUrl = `${baseUrl}/web/assets/${path}`;

    const url = service.getAssetUrl(path);

    expect(url).toBe(expectedUrl);
  });

  it("should generate correct asset API URL for download", () => {
    const path = "test-image.jpg";
    const expectedUrl = `${environment.apiUrl}/assets?path=${encodeURIComponent(
      path
    )}`;

    const url = service.getAssetApiUrl(path);

    expect(url).toBe(expectedUrl);
  });

  it("should handle special characters in asset path for API URL", () => {
    const path = "folder/test image (1).jpg";
    const encodedPath = encodeURIComponent(path);
    const expectedUrl = `${environment.apiUrl}/assets?path=${encodedPath}`;

    const url = service.getAssetApiUrl(path);

    expect(url).toBe(expectedUrl);
  });

  it("should download asset as blob", (done) => {
    const path = "test-image.jpg";
    const mockBlob = new Blob(["image data"], { type: "image/jpeg" });

    service.downloadAsset(path).subscribe((blob) => {
      expect(blob).toEqual(mockBlob);
      expect(blob.type).toBe("image/jpeg");
      done();
    });

    const expectedUrl = service.getAssetApiUrl(path);
    const req = httpMock.expectOne(expectedUrl);
    expect(req.request.method).toBe("GET");
    expect(req.request.responseType).toBe("blob");

    req.flush(mockBlob);
  });

  it("should handle upload errors", (done) => {
    const mockFile = new File(["test"], "test.jpg", { type: "image/jpeg" });

    service.uploadAsset(mockFile).subscribe({
      next: () => fail("should have failed"),
      error: (error) => {
        expect(error.status).toBe(500);
        done();
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/assets`);
    req.flush("Upload failed", {
      status: 500,
      statusText: "Internal Server Error",
    });
  });

  it("should handle batch upload errors", (done) => {
    const mockFiles = [
      new File(["content"], "file.jpg", { type: "image/jpeg" }),
    ];

    service.uploadAssetsBatch(mockFiles).subscribe({
      next: () => fail("should have failed"),
      error: (error) => {
        expect(error.status).toBe(400);
        done();
      },
    });

    const req = httpMock.expectOne(`${environment.apiUrl}/assets/batch`);
    req.flush("Bad request", { status: 400, statusText: "Bad Request" });
  });

  it("should handle download errors", (done) => {
    const path = "nonexistent.jpg";

    service.downloadAsset(path).subscribe({
      next: () => fail("should have failed"),
      error: (error) => {
        expect(error.status).toBe(404);
        done();
      },
    });

    const req = httpMock.expectOne(service.getAssetApiUrl(path));
    // For blob responses, need to pass a Blob for error responses too
    req.flush(new Blob(["Not found"], { type: "text/plain" }), {
      status: 404,
      statusText: "Not Found",
    });
  });
});
