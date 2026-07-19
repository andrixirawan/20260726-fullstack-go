"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { ArrowLeft, Send, Save, X, Upload, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { blogApi, Category, Tag, PostStatus, ApiError } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export default function NewPostPage() {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [categories, setCategories] = useState<Category[]>([])
  const [tags, setTags] = useState<Tag[]>([])

  const [title, setTitle] = useState("")
  const [excerpt, setExcerpt] = useState("")
  const [content, setContent] = useState("")
  const [coverImage, setCoverImage] = useState("")
  const [isUploadingCover, setIsUploadingCover] = useState(false)
  const [categoryId, setCategoryId] = useState("")
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([])
  const [status, setStatus] = useState<PostStatus>("draft")

  const handleCoverUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setIsUploadingCover(true)
    try {
      const result = await blogApi.uploadImage(file)
      const coverUrl = `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}${result.url}`
      setCoverImage(coverUrl)
      toast.success("Cover image uploaded successfully!")
    } catch {
      toast.error("Failed to upload cover image")
    } finally {
      setIsUploadingCover(false)
    }
  }

  useEffect(() => {
    blogApi.listCategories().then(setCategories).catch(() => {})
    blogApi.listTags().then(setTags).catch(() => {})
  }, [])

  const toggleTag = (id: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id],
    )
  }

  const handleSubmit = async (submitStatus: PostStatus) => {
    if (!title.trim()) {
      toast.error("Title is required")
      return
    }
    if (!content.trim()) {
      toast.error("Content is required")
      return
    }

    setIsSubmitting(true)
    try {
      const post = await blogApi.createPost({
        title,
        excerpt,
        content,
        cover_image: coverImage,
        category_id: categoryId || undefined,
        tag_ids: selectedTagIds.length > 0 ? selectedTagIds : undefined,
        status: submitStatus,
      })
      toast.success(
        submitStatus === "published" ? "Post published successfully!" : "Post saved as draft",
      )
      router.push(`/blog/${post.id}`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to create post")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          id="btn-back-new"
          variant="ghost"
          size="icon"
          onClick={() => router.back()}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">New Post</h1>
          <p className="text-sm text-muted-foreground">Write and publish your blog post</p>
        </div>
        <div className="ml-auto flex gap-2">
          <Button
            id="btn-save-draft"
            variant="outline"
            onClick={() => handleSubmit("draft")}
            disabled={isSubmitting}
            className="gap-2"
          >
            <Save className="h-4 w-4" />
            Save Draft
          </Button>
          <Button
            id="btn-publish-post"
            onClick={() => handleSubmit("published")}
            disabled={isSubmitting}
            className="gap-2"
          >
            <Send className="h-4 w-4" />
            Publish
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-5">
          <div className="space-y-2">
            <Label htmlFor="input-title">Title *</Label>
            <Input
              id="input-title"
              placeholder="Enter a compelling title..."
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="text-lg font-medium h-12"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="input-excerpt">Excerpt</Label>
            <Textarea
              id="input-excerpt"
              placeholder="A short summary shown in post cards (max 500 chars)..."
              value={excerpt}
              onChange={(e) => setExcerpt(e.target.value)}
              rows={3}
              maxLength={500}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground text-right">
              {excerpt.length}/500
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="input-content">Content * (Markdown supported)</Label>
            <Textarea
              id="input-content"
              placeholder="Write your post content here...

## Heading

Your content with **bold**, *italic*, `code`, and more.

```js
const hello = 'world'
```"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={20}
              className="font-mono text-sm resize-none"
            />
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-5">
          <div className="border rounded-xl p-4 bg-card space-y-5">
            <h3 className="font-semibold text-sm">Post Settings</h3>

            <div className="space-y-2">
              <Label htmlFor="input-cover-image">Cover Image (Optional)</Label>
              <div className="flex gap-2">
                <Input
                  id="input-cover-image"
                  placeholder="https://... or upload"
                  value={coverImage}
                  onChange={(e) => setCoverImage(e.target.value)}
                  disabled={isUploadingCover}
                  className="flex-1"
                />
                <div className="relative">
                  <Input
                    type="file"
                    accept="image/*"
                    onChange={handleCoverUpload}
                    disabled={isUploadingCover}
                    className="hidden"
                    id="cover-upload-file"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={() => document.getElementById("cover-upload-file")?.click()}
                    disabled={isUploadingCover}
                    title="Upload cover image"
                  >
                    {isUploadingCover ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Upload className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>
              {coverImage && (
                <div className="relative rounded-lg overflow-hidden h-28 bg-muted mt-2 group/cover">
                  <img
                    src={coverImage}
                    alt="Cover preview"
                    className="w-full h-full object-cover"
                    onError={(e) => {
                      ;(e.target as HTMLImageElement).style.display = "none"
                    }}
                  />
                  <Button
                    type="button"
                    variant="destructive"
                    size="icon"
                    className="absolute top-1.5 right-1.5 h-6 w-6 opacity-0 group-hover/cover:opacity-100 transition-opacity"
                    onClick={() => setCoverImage("")}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="select-new-category">Category</Label>
              <Select value={categoryId} onValueChange={(val) => setCategoryId(val ?? "")}>
                <SelectTrigger id="select-new-category">
                  <SelectValue placeholder="Select a category" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">No category</SelectItem>
                  {categories.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Tags</Label>
              <div className="flex flex-wrap gap-2">
                {tags.map((t) => {
                  const selected = selectedTagIds.includes(t.id)
                  return (
                    <button
                      key={t.id}
                      id={`tag-toggle-${t.id}`}
                      type="button"
                      onClick={() => toggleTag(t.id)}
                      className={`text-xs px-2.5 py-1 rounded-full border transition-colors cursor-pointer ${
                        selected
                          ? "bg-primary text-primary-foreground border-primary"
                          : "bg-transparent text-muted-foreground border-border hover:border-primary hover:text-foreground"
                      }`}
                    >
                      {t.name}
                    </button>
                  )
                })}
                {tags.length === 0 && (
                  <p className="text-xs text-muted-foreground">No tags available</p>
                )}
              </div>
            </div>
          </div>

          {/* Selected summary */}
          {selectedTagIds.length > 0 && (
            <div className="border rounded-xl p-4 bg-card space-y-2">
              <h3 className="font-semibold text-sm">Selected Tags</h3>
              <div className="flex flex-wrap gap-1">
                {selectedTagIds.map((id) => {
                  const tag = tags.find((t) => t.id === id)
                  return tag ? (
                    <Badge
                      key={id}
                      variant="secondary"
                      className="gap-1 cursor-pointer"
                      onClick={() => toggleTag(id)}
                    >
                      {tag.name}
                      <X className="h-3 w-3" />
                    </Badge>
                  ) : null
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
